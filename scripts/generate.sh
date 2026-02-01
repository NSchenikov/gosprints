#!/bin/bash

echo "=== Генерация Go кода из proto файлов ==="

# Проверяем наличие protoc
if ! command -v protoc &> /dev/null; then
    echo "protoc не установлен. Установите:"
    echo "   brew install protobuf"
    exit 1
fi

# Проверяем наличие Go плагинов
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Установка protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Установка protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Проверяем proto файл
if [ ! -f "api/proto/task/v1/task.proto" ]; then
    echo "Proto файл не найден: api/proto/task/v1/task.proto"
    exit 1
fi

echo "Генерация кода из proto файла..."

# Генерируем код
protoc --proto_path=api/proto/task/v1 \
       --go_out=internal/grpc/task/pb \
       --go_opt=paths=source_relative \
       --go-grpc_out=internal/grpc/task/pb \
       --go-grpc_opt=paths=source_relative \
       api/proto/task/v1/task.proto

if [ $? -eq 0 ]; then
    echo "Код успешно сгенерирован в internal/grpc/task/pb/"
    echo "Созданные файлы:"
    ls -la internal/grpc/task/pb/
else
    echo "Ошибка генерации кода"
    exit 1
fi
