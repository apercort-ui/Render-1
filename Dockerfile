FROM golang:1.22-alpine AS builder
WORKDIR /app

# Копируем исправленный go.mod
COPY go.mod ./

# Теперь tidy создаст идеальный go.sum, так как модули жестко прописаны
RUN go mod tidy

# Копируем код и компилируем
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o site-core .

# Финальный контейнер
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/site-core .
EXPOSE 8080
CMD ["./site-core"]
