файлы БД для postgresql можно взять тут https://disk.yandex.ru/d/sa3wbixqpbejRA

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

ждем sleep 10

docker run -d --name kafka \
 --platform linux/amd64 \
 --network mentoring-net \
 -p 9092:9092 \
 -e KAFKA_BROKER_ID=1 \
 -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
 -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
 -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
 confluentinc/cp-kafka:6.2.0

ждем sleep 20

# проверяем работоспособность

docker ps

# Создаем топик

docker exec kafka kafka-topics --create --topic task-events --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

# еще раз проверяем, должеть быть task-events

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

---

# РАБОТА API

# 1. Регистрация и логин

curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"user1","password":"pass"}'
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"user1","password":"pass"}'
TOKEN="ваш_токен"

# 2. Создать задачу

curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"text":"Купить молоко"}'

# 3. Получить задачи

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/tasks

# 4. Поиск задач

curl -H "Authorization: Bearer $TOKEN" "http://localhost:8080/tasks/search?q=молоко"

# 5. WebSocket (в отдельном терминале)

wscat -c "ws://localhost:8082/ws?user_id=user1"
