package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

var (
	ctx         = context.Background()
	redisClient *redis.ClusterClient // или *redis.Client, для Upstash обычно используется стандартный
	rdb         *redis.Client
	rabbitConn  *amqp.Connection
)

func main() {
	// 1. Получаем настройки из переменных окружения (Render + Сторонние сервисы)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Локальный порт для тестов
	}

	redisURL := os.Getenv("REDIS_URL")
	rabbitURL := os.Getenv("RABBITMQ_URL")

	log.Printf("Запуск ядра сайта на порту %s...", port)

	// 2. Инициализация внешнего Redis (например, Upstash)
	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Ошибка парсинга REDIS_URL: %v", err)
		} else {
			rdb = redis.NewClient(opt)
			// Проверяем подключение
			_, err := rdb.Ping(ctx).Result()
			if err != nil {
				log.Printf("Не удалось подключиться к Redis: %v", err)
			} else {
				log.Println("Успешное подключение к Redis!")
			}
		}
	} else {
		log.Println("Предупреждение: REDIS_URL не задан. Работа без кэша.")
	}

	// 3. Инициализация RabbitMQ (например, CloudAMQP)
	if rabbitURL != "" {
		var err error
		rabbitConn, err = amqp.Dial(rabbitURL)
		if err != nil {
			log.Printf("Не удалось подключиться к RabbitMQ: %v", err)
		} else {
			log.Println("Успешное подключение к RabbitMQ!")
			defer rabbitConn.Close()
		}
	} else {
		log.Println("Предупреждение: RABBITMQ_URL не задан. Очереди отключены.")
	}

	// 4. Настройка маршрутов (Эндпоинтов сайта)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)

	// 5. Запуск веб-сервера
	serverAddr := fmt.Sprintf(":%s", port)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// Главная страница сайта
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Пример работы с Redis (счетчик просмотров)
	visits := "недоступно"
	if rdb != nil {
		val, err := rdb.Incr(ctx, "site_visits").Result()
		if err == nil {
			visits = fmt.Sprintf("%d", val)
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<h1>Добро пожаловать на сайт!</h1><p>Ядро на Go работает внутри Docker.</p><p>Просмотров главной страницы через Redis: %s</p>", visits)
}

// Проверка работоспособности для Render (Health Check)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}


