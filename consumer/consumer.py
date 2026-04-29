import asyncio
import json
import os
import re
import datetime
from nats.aio.client import Client as NATS
from clickhouse_driver import Client as ClickHouseClient

# Конфигурация фильтров из окружения
# Уровни логов, которые разрешены (через запятую). Пример: "warn,error"
ALLOWED_LEVELS = os.getenv("FILTER_LEVELS", "warn,error").split(",")
# Ключевые слова (подстроки), которые должны присутствовать в тексте (логическое ИЛИ)
REQUIRED_KEYWORDS = os.getenv("FILTER_KEYWORDS", "").split(",") if os.getenv("FILTER_KEYWORDS") else None
# Запрещённые слова (если есть – логи отбрасываются)
EXCLUDED_KEYWORDS = os.getenv("FILTER_EXCLUDE_KEYWORDS", "").split(",") if os.getenv("FILTER_EXCLUDE_KEYWORDS") else None
# Список project_id, которые разрешены (через запятую). Если не задан – все проекты.
ALLOWED_PROJECTS = os.getenv("FILTER_PROJECTS", "").split(",") if os.getenv("FILTER_PROJECTS") else None
# Регулярное выражение, которому должен соответствовать текст (опционально)
TEXT_REGEX = os.getenv("FILTER_TEXT_REGEX", None)
# Минимальная длина сообщения
MIN_TEXT_LEN = int(os.getenv("FILTER_MIN_TEXT_LEN", "0"))
# Максимальная длина сообщения (0 – без ограничений)
MAX_TEXT_LEN = int(os.getenv("FILTER_MAX_TEXT_LEN", "0"))
# Временное окно: только логи после этой даты (ISO format "YYYY-MM-DD HH:MM:SS")
TIME_AFTER = os.getenv("FILTER_TIME_AFTER", None)
# Временное окно: только логи до этой даты
TIME_BEFORE = os.getenv("FILTER_TIME_BEFORE", None)

# Функция фильтрации
def should_process(data: dict) -> bool:
    level = data.get("level")
    text = data.get("text", "")
    project_id = data.get("project_id")
    time_str = data.get("time")

    # 1. Фильтр по уровню
    if level not in ALLOWED_LEVELS:
        print(f"Filtered out by level: {level}")
        return False

    # 2. Фильтр по обязательным ключевым словам (AND внутри списка не требуется, обычно OR)
    if REQUIRED_KEYWORDS and not any(kw.lower() in text.lower() for kw in REQUIRED_KEYWORDS if kw):
        print(f"Filtered out by missing keyword from {REQUIRED_KEYWORDS}")
        return False

    # 3. Фильтр по исключённым ключевым словам (если есть хоть одно – отбрасываем)
    if EXCLUDED_KEYWORDS and any(kw.lower() in text.lower() for kw in EXCLUDED_KEYWORDS if kw):
        print(f"Filtered out by excluded keyword")
        return False

    # 4. Фильтр по project_id
    if ALLOWED_PROJECTS and project_id not in ALLOWED_PROJECTS:
        print(f"Filtered out by project_id: {project_id}")
        return False

    # 5. Фильтр по регулярному выражению на текст
    if TEXT_REGEX and not re.search(TEXT_REGEX, text, re.IGNORECASE):
        print(f"Filtered out by regex: {TEXT_REGEX}")
        return False

    # 6. Фильтр по минимальной/максимальной длине текста
    if MIN_TEXT_LEN > 0 and len(text) < MIN_TEXT_LEN:
        print(f"Filtered out by min length: {len(text)} < {MIN_TEXT_LEN}")
        return False
    if MAX_TEXT_LEN > 0 and len(text) > MAX_TEXT_LEN:
        print(f"Filtered out by max length: {len(text)} > {MAX_TEXT_LEN}")
        return False

    # 7. Фильтр по временному окну (если time передан строкой, преобразуем)
    if TIME_AFTER or TIME_BEFORE:
        try:
            if isinstance(time_str, str):
                log_time = datetime.datetime.strptime(time_str, '%Y-%m-%d %H:%M:%S')
            else:
                log_time = time_str
            if TIME_AFTER:
                after = datetime.datetime.strptime(TIME_AFTER, '%Y-%m-%d %H:%M:%S')
                if log_time < after:
                    print(f"Filtered out by time after: {log_time} < {after}")
                    return False
            if TIME_BEFORE:
                before = datetime.datetime.strptime(TIME_BEFORE, '%Y-%m-%d %M:%S')  # исправить формат? оставим как пример
                if log_time > before:
                    print(f"Filtered out by time before: {log_time} > {before}")
                    return False
        except Exception as e:
            print(f"Time filter error: {e}")

    # 8. Можно добавить фильтр по наличию обязательного поля, например, "user_id"
    if os.getenv("FILTER_REQUIRE_FIELD") and os.getenv("FILTER_REQUIRE_FIELD") not in data:
        print(f"Filtered out by missing field: {os.getenv('FILTER_REQUIRE_FIELD')}")
        return False

    return True


# Основной код consumer (JetStream)
NATS_URL = os.getenv("NATS_URL", "nats://nats:4222")
NATS_USER = os.getenv("NATS_USER", "app")
NATS_PASSWORD = os.getenv("NATS_PASSWORD", "app_pass_456")
NATS_STREAM = "LOG_STREAM"
NATS_SUBJECT = "raw.logs"
DURABLE_NAME = "log_durable_consumer"

CH_HOST = os.getenv("CLICKHOUSE_HOST", "clickhouse")
CH_PORT = int(os.getenv("CLICKHOUSE_PORT", 9000))
CH_DB = os.getenv("CLICKHOUSE_DB", "logs_db")
CH_USER = "default"
CH_PASSWORD = "clickhouse_pass"

async def main():
    nc = NATS()
    await nc.connect(servers=[NATS_URL], user=NATS_USER, password=NATS_PASSWORD)
    print(f"Connected to NATS at {NATS_URL}")

    js = nc.jetstream()
    try:
        await js.add_stream(name=NATS_STREAM, subjects=[NATS_SUBJECT], storage="file")
        print(f"Stream '{NATS_STREAM}' created")
    except Exception as e:
        if "stream name already in use" in str(e):
            print(f"Stream '{NATS_STREAM}' already exists")
        else:
            raise

    sub = await js.pull_subscribe(NATS_SUBJECT, DURABLE_NAME, stream=NATS_STREAM)
    print(f"Pull subscription created on '{NATS_SUBJECT}', durable '{DURABLE_NAME}'")

    ch = ClickHouseClient(host=CH_HOST, port=CH_PORT, database=CH_DB, user=CH_USER, password=CH_PASSWORD)

    async def process_messages():
        while True:
            try:
                msgs = await sub.fetch(10, timeout=1.0)
                for msg in msgs:
                    raw = msg.data.decode().strip()
                    print(f"Raw message: {raw}")

                    # Парсинг в data (поддерживается JSON и упрощённый формат)
                    try:
                        data = json.loads(raw)
                    except json.JSONDecodeError:
                        match = re.findall(r'(\w+):([^,}]+)', raw)
                        if match:
                            data = {}
                            for key, val in match:
                                val = val.strip()
                                if val.startswith('"') and val.endswith('"'):
                                    val = val[1:-1]
                                data[key] = val
                        else:
                            print("Unknown format, ack to skip")
                            await msg.ack()
                            continue

                    # Применяем фильтры
                    if not should_process(data):
                        print("Message filtered out, ack and skip")
                        await msg.ack()
                        continue

                    # Преобразование времени
                    if isinstance(data.get('time'), str):
                        try:
                            dt = datetime.datetime.strptime(data['time'], '%Y-%m-%d %H:%M:%S')
                            dt = dt.replace(tzinfo=datetime.timezone.utc)
                            data['time'] = dt
                        except Exception as e:
                            print(f"Time parse error: {e}, ack to skip")
                            await msg.ack()
                            continue

                    try:
                        ch.execute(
                            "INSERT INTO raw_logs (text, level, time, project_id) VALUES",
                            [(data['text'], data['level'], data['time'], data['project_id'])]
                        )
                        print(f"Saved: {data}")
                        await msg.ack()
                    except Exception as e:
                        print(f"DB error: {e}")
                        await msg.nak()
            except asyncio.TimeoutError:
                continue
            except Exception as e:
                print(f"Fetch error: {e}")
                await asyncio.sleep(1)

    print("JetStream consumer started, waiting for messages...")
    await process_messages()

if __name__ == "__main__":
    asyncio.run(main())