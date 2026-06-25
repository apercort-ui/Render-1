package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"        // Драйвер Postgres (импорт ради побочного эффекта)
	"github.com/streadway/amqp" // Драйвер RabbitMQ
)

// Глобальная переменная-указатель для базы данных. 
// Звёздочка означает, что здесь хранится адрес в памяти, а не сама база.
var db *sql.DB

// Структура для десериализации JSON от TypeScript
type FrontendMessage struct {
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Инициализируем подключение к базе данных Neon Postgres
	initDB()
	// defer гарантирует, что пул соединений закроется только при остановке сервера
	if db != nil {
		defer db.Close()
	}

	mux := http.NewServeMux()

	
// Раздаем статику: index.html лежит в корне frontend, а скрипты в dist
	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fileServer)
	mux.Handle("/dist/", fileServer)
	// Эндпоинты API
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/send", handleSendMessage)

	log.Printf("Сервер запущен на порту %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// Функция инициализации подключения к Postgres
func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("[Предупреждение] Переменная DATABASE_URL не настроена. Шаг пропускается.")
		return
	}

	// 1. Устанавливаем конфигурацию подключения (sql.Open не создает сетевое соединение сразу)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("[Ошибка] Не удалось настроить драйвер базы данных: %v", err)
		return
	}

	// 2. Реальная проверка связи через Указатель.
	// Функция Ping() отправляет короткий запрос бронепоезду Neon, чтобы проверить, жива ли база.
	err = db.Ping()
	if err != nil {
		log.Printf("[Ошибка] База данных Neon недоступна через Ping: %v", err)
		// На всякий случай обнуляем указатель, если связи нет
		db = nil 
		return
	}

	log.Println("[Успех] Успешное подключение к Neon PostgreSQL! Проверка Ping пройдена.")
}

// Обработчик отправки сообщений в RabbitMQ
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var msg FrontendMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("Ошибка JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if msg.Text == "" {
		http.Error(w, "Сообщение пустое", http.StatusBadRequest)
		return
	}

	log.Printf("Отправка сообщения в RabbitMQ: %s", msg.Text)
	err = publishToRabbitMQ(msg.Text)
	if err != nil {
		log.Printf("Ошибка RabbitMQ: %v", err)
		http.Error(w, "Ошибка отправки в очередь", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Добавлено в очередь"}`))
}

// Функция взаимодействия с AMQP брокером
func publishToRabbitMQ(messageText string) error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		return errors.New("переменная RABBITMQ_URL не настроена")
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("jobs_queue", false, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(messageText),
	})
	return err
}

// Обновленный обработчик статуса, сообщающий фронтенду о состоянии Postgres
func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	hasRabbit := "false"
	if os.Getenv("RABBITMQ_URL") != "" {
		hasRabbit = "true"
	}

	// Проверяем, инициализирован ли наш глобальный указатель базы данных
	hasPostgres := "false"
	if db != nil {
		hasPostgres = "true"
	}
	
	w.Write([]byte(`{
		"status": "online",
		"redisConnected": false,
		"rabbitmqConnected": ` + hasRabbit + `,
		"postgresConnected": ` + hasPostgres + `,
		"timestamp": "now"
	}`))
}
