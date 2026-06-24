FROM golang:1.22-alpine AS builder
WORKDIR /app

# 1. Сначала копируем И код, И описание модуля
COPY go.mod main.go ./

# 2. Теперь tidy увидит импорты внутри main.go и всё скачает!
RUN go mod tidy

# 3. Собираем проект
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o site-core .

# Финальный контейнер
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/site-core .
EXPOSE 8080
CMD ["./site-core"]
