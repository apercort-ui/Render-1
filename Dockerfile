FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY . .
RUN npm install -g typescript && tsc frontend/app.ts --outDir frontend/dist --target es6

FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o site-core .

FROM alpine:latest
WORKDIR /root/
COPY --from=backend-builder /app/site-core .
COPY --from=frontend-builder /app/frontend/ ./frontend/
CMD ["./site-core"]
