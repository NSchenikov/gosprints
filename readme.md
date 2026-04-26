файлы БД для postgresql можно взять тут https://disk.yandex.ru/d/sa3wbixqpbejRA
Для работы state machine добавить в базу:
ALTER TABLE "Tasks" ADD COLUMN attempts INT DEFAULT 0;
ALTER TABLE "Tasks" ADD COLUMN validation1_at TIMESTAMP;
ALTER TABLE "Tasks" ADD COLUMN validation2_at TIMESTAMP;
ALTER TABLE "Tasks" ADD COLUMN closed_at TIMESTAMP;

ЗАПУСК ПРОЕКТА:
Установить и запустить Docker

В 1 терминале:
переходим в корневую папку (можно использовать cd)

Остановить и удалить старые контейнеры с помощью docker rm -f kafka zookeeper

# Запускаем PostgreSQL (если не запущен)

docker run -d --name postgres \
 -p 8000:5432 \
 -e POSTGRES_DB=gosprints \
 -e POSTGRES_USER=postgres \
 -e POSTGRES_PASSWORD=4840707101 \
 postgres:15

# создаем сеть

docker network create mentoring-net

# Запускаем Zookeeper и Kafka в сети

docker run -d --name zookeeper \
 --platform linux/amd64 \
 --network mentoring-net \
 -p 2181:2181 \
 -e ZOOKEEPER_CLIENT_PORT=2181 \
 confluentinc/cp-zookeeper:6.2.0

ждем
sleep 10

docker run -d --name kafka \
 --platform linux/amd64 \
 --network mentoring-net \
 -p 9092:9092 \
 -e KAFKA_BROKER_ID=1 \
 -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
 -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
 -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
 confluentinc/cp-kafka:6.2.0

ждем
sleep 20

# проверяем работоспособность

docker ps

# Создаем топик

docker exec kafka kafka-topics --create --topic task-events --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

# еще раз проверяем, должен быть task-events

docker exec kafka kafka-topics --list --bootstrap-server 127.0.0.1:9092

# 1) запускаем первый терминал (task-service)

cd task-service
export KAFKA_BROKERS=127.0.0.1:9092
export KAFKA_TOPIC=task-events
go run ./cmd/main.go

# 2) запускаем второй терминал (notification-service)

cd notification-service
export KAFKA_BROKERS=127.0.0.1:9092
export KAFKA_TOPIC=task-events
export KAFKA_GROUP_ID=notification-group
go run ./cmd/main.go

# 3) Запускаем третий терминал (api-gateway)

cd api-gateway && go run ./cmd/main.go

# 4) Запуск ETL

при необходимости остановить и удалить существующий контейнер
docker rm -f clickhouse

запустить (clickhouse работает с паролем)
docker run -d --name clickhouse \
 -p 8123:8123 \
 -p 9000:9000 \
 -e CLICKHOUSE_USER=default \
 -e CLICKHOUSE_PASSWORD=clickhouse \
 clickhouse/clickhouse-server:latest

Проверяем:
docker ps | grep clickhouse

Подключаемся и открываем консоль clickhouse:
docker exec -it clickhouse clickhouse-client

# 5) Для создания базы и аналитики (если нужно) в консоли clickhouse выполнить

CREATE DATABASE IF NOT EXISTS analytics;
USE analytics;

CREATE TABLE IF NOT EXISTS task_analytics (
user_id String,
tasks_completed Int32,
avg_completion_time Float64,
last_event_time DateTime,
date Date
) ENGINE = ReplacingMergeTree(last_event_time)
ORDER BY (user_id);

Затем выйти из консоли (ctrl+D)

переходим
cd etl-worker

загружаем переменные из .env
source .env

запускаем etl-worker
go run ./cmd/main.go

---

# РАБОТА API

# 1. Регистрация и логин

curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"user1","password":"pass"}'
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"user1","password":"pass"}'
TOKEN="ваш_токен"

# 2. Создать задачу (или лучше даже несколько для наглядности тестирования аналитики)

curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"text":"Купить молоко"}'

# проверить событие (аналитику) в clickhouse

docker exec -it clickhouse clickhouse-client --query "
SELECT user_id, tasks_completed, avg_completion_time
FROM analytics.task_analytics FINAL

# 3. Получить задачи

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/tasks

# 4. Поиск задач

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/tasks/search?q=молоко"

# 5. WebSocket (в отдельном терминале)

wscat -c "ws://localhost:8082/ws?user_id=user1"
