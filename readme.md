файлы БД для postgresql можно взять тут https://disk.yandex.ru/d/sa3wbixqpbejRA

Регистрация curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{"username":"имя пользователя", "password":"пароль"}'

Получить токен
curl -X POST http://localhost:8080/login -H "Content-Type: application/json" -d '{"username":"имя пользователя","password":"пароль"}'

Использовать токен для доступа:
Заменить YOUR_TOKEN на полученный токен

для отслеживания статуса задачи по WebSocket установить wscat и использовать:
wscat -c "ws://localhost:8080/ws" \
 -H "Authorization: Bearer $YOUR_TOKEN"

Все задачи: curl -H "Authorization: Bearer $YOUR_TOKEN" http://localhost:8080/tasks

Добавить задачу:
curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer $YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"текст задачи"}'

Прочитать задачу по id:
curl -H "Authorization: Bearer $YOUR_TOKEN" http://localhost:8080/tasks/{id}

Обновить задачу:
curl -X PUT http://localhost:8080/tasks/{id} -H "Authorization: Bearer $YOUR_TOKEN" -H "Content-Type: application/json" -d '{"text":"Новый текст"}'

Удалить задачу:
curl -X DELETE -H "Authorization: Bearer $YOUR_TOKEN" http://localhost:8080/tasks/{id}

//КЭШИРОВАНИЕ:

# 1. Очистите кэш (на всякий случай)

curl -X POST http://localhost:8080/admin/cache/clear \
 -H "Authorization: Bearer $YOUR_TOKEN"

# 2. Проверьте - должна быть 0 статистика

curl -X GET http://localhost:8080/admin/cache/stats \
 -H "Authorization: $Bearer YOUR_TOKEN"

# 3. Сделайте 3 запроса к API

echo "=== Тест API кэширования ==="

echo "Запрос 1 (должен быть MISS):"
time curl -s -X GET "http://localhost:8080/tasks" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

echo "Запрос 2 (должен быть HIT, быстрее):"
time curl -s -X GET "http://localhost:8080/tasks" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

echo "Запрос 3 (должен быть HIT, еще быстрее):"
time curl -s -X GET "http://localhost:8080/tasks" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

# 4. Проверьте статистику

echo "=== Статистика после 3 запросов ==="
curl -X GET http://localhost:8080/admin/cache/stats \
 -H "Authorization: Bearer $YOUR_TOKEN"

# ОЖИДАЕМ:

# misses=1 (первый запрос)

# hits=2 (второй и третий запросы)

# sets=1 (сохранение после первого запроса)

# hit_rate=66.7%

//Инвалидация

# 1. Создайте новую задачу (должна инвалидировать кэш списка)

echo "=== Создание новой задачи ==="
curl -X POST http://localhost:8080/tasks \
 -H "Authorization: Bearer $YOUR_TOKEN" \
 -H "Content-Type: application/json" \
 -d '{"text": "Тест инвалидации кэша"}'

# 2. Сразу запросите список задач

echo "=== Запрос списка после создания ==="
time curl -s -X GET "http://localhost:8080/tasks" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

# 3. Проверьте статистику

echo "=== Статистика после инвалидации ==="
curl -X GET http://localhost:8080/admin/cache/stats \
 -H "Authorization: Bearer $YOUR_TOKEN"

# ОЖИДАЕМ:

# misses увеличилось на 1 (кэш был инвалидирован)

# sets увеличилось на 1 (новые данные сохранены в кэш)

//тест конкретной задачи

# Получите ID созданной задачи (например, 27)

TASK_ID=27

echo "=== Тест кэширования конкретной задачи ==="

echo "Запрос задачи $TASK_ID (первый раз - MISS):"
time curl -s -X GET "http://localhost:8080/tasks/$TASK_ID" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

echo "Запрос задачи $TASK_ID (второй раз - HIT, быстрее):"
time curl -s -X GET "http://localhost:8080/tasks/$TASK_ID" \
 -H "Authorization: Bearer $YOUR_TOKEN" > /dev/null

echo "=== Финальная статистика ==="
curl -X GET http://localhost:8080/admin/cache/stats \
 -H "Authorization: Bearer $YOUR_TOKEN"

//МЕТРИКИ

# 1. Установить jq

brew install jq

# 2. Получаем метрики в JSON

curl http://localhost:8080/metrics | jq '.'

# 3. Получаем Prometheus метрики

curl http://localhost:8080/metrics/prometheus

# 4. Проверяем обновление метрик

# Создаем задачу

curl -X POST http://localhost:8080/tasks \
 -H "Authorization: Bearer $YOUR_TOKEN" \
 -d '{"text": "Test"}'

# Проверяем метрики снова

curl http://localhost:8080/metrics | jq '.tasks'

//gRPC
brew install protobuf //установка
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

which protoc-gen-go
which protoc-gen-go-grpc

# Проверьте что protoc установился

protoc --version

# Создайте директорию для сгенерированных файлов

mkdir -p internal/grpc/task/pb

# Сгенерируйте код

protoc -I api/proto/task/v1 \
 --go_out=internal/grpc/task/pb \
 --go_opt=paths=source_relative \
 --go-grpc_out=internal/grpc/task/pb \
 --go-grpc_opt=paths=source_relative \
 api/proto/task/v1/task.proto

# Проверьте

ls -la internal/grpc/task/pb/
