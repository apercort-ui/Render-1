FROM golang:1.22-alpine AS builder
WORKDIR /app

# Копируем go.mod
COPY go.mod ./

# Запускаем tidy и выводим структуру окружения в логи
RUN go mod tidy && echo "=== GO MOD OK ===" && ls -la

# Копируем main.go
COPY main.go ./

# Проверяем синтаксис main.go без сборки
RUN go vet main.go
