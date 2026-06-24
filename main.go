package main

import (
	"log"
	"net/http"
	"os"
	// Ваши остальные импорты (например, github.com/streadway/amqp и т.д.) здесь, если нужны
)

// Middleware для настройки CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обработка предварительного запроса браузера (Preflight OPTIONS)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// ... ваша логика инициализации Redis и RabbitMQ ...

	// 1. Создаем маршрутизатор
	mux := http.NewServeMux()

	// 2. Регистрируем API эндпоинт
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "live",
			"redisConnected": false,
			"rabbitmqConnected": true,
			"timestamp": "2026-06-24T18:23:53Z"
		}`))
	})

	// 3. Указываем Go раздавать статические HTML/CSS/JS файлы из папки "./frontend"
	// ВАЖНО: Эта строчка должна быть ВНЕ блоков ошибок и ДО ListenAndServe
	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fileServer)

	// 4. Определяем порт окружения Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	log.Printf("Запуск ядра сайта на порту %s...", port)

	// 5. Запускаем сервер (оборачиваем маршруты в CORS)
	err := http.ListenAndServe(":"+port, corsMiddleware(mux))
	if err != nil {
		log.Fatalf("Критическая ошибка запуска сервера: %v", err)
	}
}
