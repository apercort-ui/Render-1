package main

module site-core

go 1.22

require (
	github.com/lib/pq v1.10.9
	github.com/streadway/amqp v1.1.0
)

// Структура для десериализации JSON от TypeScript
type FrontendMessage struct {
	Text string `json:"text"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Раздача статичных файлов фронтенда
	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fileServer)

	// Эндпоинты API
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/send", handleSendMessage)

	log.Printf("Сервер запущен на порту %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
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
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if msg.Text == "" {
		http.Error(w, "Сообщение не может быть пустым", http.StatusBadRequest)
		return
	}

	log.Printf("Отправка сообщения в RabbitMQ: %s", msg.Text)
	err = publishToRabbitMQ(msg.Text)
	if err != nil {
		log.Printf("Ошибка RabbitMQ: %v", err)
		http.Error(w, "Ошибка отправки в очередь брокера", http.StatusInternalServerError)
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

	q, err := ch.QueueDeclare(
		"jobs_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(messageText),
		},
	)
	return err
}

// Полностью исправленный обработчик статуса без os.String()
func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	hasRabbit := "false"
	if os.Getenv("RABBITMQ_URL") != "" {
		hasRabbit = "true"
	}
	
	w.Write([]byte(`{
		"status": "online",
		"redisConnected": false,
		"rabbitmqConnected": ` + hasRabbit + `,
		"timestamp": "now"
	}`))
}
