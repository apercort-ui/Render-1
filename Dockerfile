# === ЭТАП 1: Сборка фронтенда (TypeScript) ===
FROM node:18-alpine AS frontend-builder
WORKDIR /app
# Копируем всё содержимое репозитория
COPY . .
# Устанавливаем TypeScript и компилируем app.ts в папку frontend/dist
RUN npm install -g typescript && tsc

# === ЭТАП 2: Сборка бэкенда (Go) ===
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod main.go ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o site-core .

# === ЭТАП 3: Финальный контейнер ===
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копируем скомпилированное Go-ядро
COPY --from=backend-builder /app/site-core .

# Копируем папку frontend (внутри которой благодаря Этапу 1 уже лежит готовый dist/app.js!)
COPY --from=frontend-builder /app/frontend/ ./frontend/

EXPOSE 8080
CMD ["./site-core"]
