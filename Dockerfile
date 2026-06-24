FROM golang:1.22-alpine AS builder
WORKDIR /app

# Копируем ТОЛЬКО go.mod (go.sum удален, его тут нет)
COPY go.mod ./

# Принудительно генерируем чистый go.sum прямо внутри Docker
RUN go mod tidy

# Копируем код и собираем проект
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o site-core .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/site-core .
EXPOSE 8080
CMD ["./site-core"]
