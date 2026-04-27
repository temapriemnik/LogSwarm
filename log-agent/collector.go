package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	docker *client.Client
	js     nats.JetStreamContext
	apiKey string
}

func NewCollector(docker *client.Client, js nats.JetStreamContext, apiKey string) *Collector {
	return &Collector{
		docker: docker,
		js:     js,
		apiKey: apiKey,
	}
}

func (c *Collector) Start(ctx context.Context) error {
	containers, err := c.docker.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	log.Printf("Found %d containers", len(containers))

	for _, container := range containers {
		go c.tailContainer(ctx, container.ID, container.Names)
	}

	<-ctx.Done()
	return nil
}

func (c *Collector) tailContainer(ctx context.Context, containerID string, names []string) {
	defer func() {
		ContainersCurrent.Dec()
	}()

	ContainersCurrent.Inc()

	logReader, err := c.docker.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
		Tail:       "",
	})
	if err != nil {
		log.Printf("Failed to get logs for container %s: %v", containerID, err)
		ErrorsTotal.With(prometheus.Labels{"type": "container_logs"}).Inc()
		return
	}
	defer logReader.Close()

	containerName := ""
	if len(names) > 0 {
		containerName = names[0]
	}

	scanner := bufio.NewReader(logReader)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := scanner.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			ErrorsTotal.With(prometheus.Labels{"type": "read_error"}).Inc()
			break
		}

		if len(line) > 8 {
			line = line[8:]
		}

		start := time.Now()

		msg := LogMessage{
			APIToken:  c.apiKey,
			Log:       containerName + ": " + line,
			Timestamp: time.Now().UTC(),
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message: %v", err)
			ErrorsTotal.With(prometheus.Labels{"type": "marshal_error"}).Inc()
			continue
		}

		_, err = c.js.Publish("raw.logs", data)
		ProcessingLatency.Observe(time.Since(start).Seconds())

		if err != nil {
			log.Printf("Failed to publish log: %v", err)
			ErrorsTotal.With(prometheus.Labels{"type": "publish_error"}).Inc()
			continue
		}

		LogsTotal.Inc()
	}
}