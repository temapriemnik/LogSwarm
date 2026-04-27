package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/client"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	APIKey  string
	NATSURL string
}

type LogMessage struct {
	APIToken  string    `json:"api_key"`
	Log       string    `json:"log"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	cfg := Config{
		APIKey:  os.Getenv("API_KEY"),
		NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),
	}

	if cfg.APIKey == "" {
		log.Fatal("API_KEY is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to get JetStream context: %v", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "logs",
		Subjects: []string{"raw.logs"},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Fatalf("Failed to create stream: %v", err)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	collector := NewCollector(dockerClient, js, cfg.APIKey)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Println("Starting HTTP server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Starting log collector...")
	if err := collector.Start(ctx); err != nil {
		log.Fatalf("Collector error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}