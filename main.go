package main

import (
	"log"
	"net/http"
	"os"
	// ваши остальные импорты (amqp, redis и т.д.)
)

// Middleware для настройки CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого источника. 
		// Для продакшена вместо "*" лучше указать конкретный URL вашего фронтенда
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Браузеры сначала отправляют предварительный запрос OPTIONS (Preflight)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// ... ваша логика инициализации Redis и RabbitMQ ...

	// Создаем маршрутизатор (или используйте ваш существующий)
	mux := http.NewServeMux()
	
	// Пример эндпоинта, который будет отдавать статус
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Отдаем JSON (в реальном коде проверяйте состояние переменных redisConnected и rabbitmqConnected)
		w.Write([]byte(`{
			"status": "live",
			"redisConnected": false,
			"rabbitmqConnected": true,
			"timestamp": "2026-06-24T18:23:53Z"
		}`))
	})

	// Определяем порт
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	log.Printf("Запуск ядра сайта на порту %s...", port)
	
	// Оборачиваем наш маршрутизатор в CORS middleware
	err := http.ListenAndServe(":"+port, corsMiddleware(mux))
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
