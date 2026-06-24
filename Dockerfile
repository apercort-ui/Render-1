# Сборка бинарного файла Go
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем описание зависимостей
COPY go.mod ./
# Если у вас появится go.sum после выполнения команд, раскомментируйте строку ниже:
# COPY go.sum ./

RUN go mod download

# Копируем исходный код
COPY main.go ./

# Собираем оптимизированный бинарник для Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o site-core .

# Финальный легковесный контейнер для запуска
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Переносим скомпилированный файл из предыдущего шага
COPY --from=builder /app/site-core .

# Экспонируем порт приложения
EXPOSE 8080

# Команда для старта контейнера
CMD ["./site-core"]
