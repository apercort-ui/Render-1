package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/streadway/amqp" // Драйвер для работы с RabbitMQ
)

// 1. Описываем структуру JSON, который к нам прилетает от TypeScript.
// Это аналог интерфейса во фронтенде. Тег `json:"text"` говорит Go, 
// какое именно поле искать в пришедшем JSON-объекте.
type FrontendMessage struct {
	Text string `json:"text"`
}

func main() {
	// Определяем порт. Render автоматически передает его в переменную окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Раздача статичных файлов фронтенда (HTML, CSS, скомпилированный JS)
	fileServer := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fileServer)

	// Эндпоинт проверки статуса систем (вызывается фронтендом каждые 10 сек)
	mux.HandleFunc("/api/status", handleStatus)

	// Эндпоинт для приема сообщений от нашей TypeScript-кнопки
	mux.HandleFunc("/api/send", handleSendMessage)

	log.Printf("Сервер успешно запущен и слушает порт %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

// 2. Функция-обработчик для эндпоинта /api/send
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Добавляем CORS-заголовки, чтобы браузер не блокировал запросы
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Если это предварительный запрос браузера (OPTIONS), сразу отвечаем успехом
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Разрешаем только POST-запросы
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Читаем JSON, пришедший из TypeScript.
	// r.Body — это поток сырых байт. Мы создаем декодер и "десериализуем" данные в структуру msg.
	var msg FrontendMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Если сообщение пустое, прерываем обработку
	if msg.Text == "" {
		http.Error(w, "Сообщение не может быть пустым", http.StatusBadRequest)
		return
	}

	// Вызываем нашу функцию отправки сообщения в RabbitMQ
	log.Printf("Получено сообщение с фронтенда: %s. Отправляем в RabbitMQ...", msg.Text)
	err = publishToRabbitMQ(msg.Text)
	if err != nil {
		log.Printf("Ошибка брокера RabbitMQ: %v", err)
		http.Error(w, "Ошибка отправки в очередь брокера", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ фронтенду в формате JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Сообщение успешно помещено в очередь"}`))
}

// 3. Функция отправки данных в RabbitMQ (то, что мы разобрали на прошлом шаге)
func publishToRabbitMQ(messageText string) error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		// Если переменная не задана, пишем заглушку (чтобы не падать по панике)
		return errors.New("RABBITMQ_URL не настроена в переменных окружения Render")
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
		"jobs_queue", // Имя нашей очереди
		false,        // Durable
		false,        // Delete when unused
		false,        // Exclusive
		false,        // No-wait
		nil,          // Arguments
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
			Body:        []byte(messageText), // Конвертируем строку в байты
		},
	)
	return err
}

// 4. Вспомогательная функция для эндпоинта /api/status
func handleStatus(w http.ResponseWriter, r *http.Header) {
	// (Стандартный код статуса, который у вас уже был в прошлых версиях)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	// Проверяем, настроена ли переменная RabbitMQ для индикации на фронте
	hasRabbit := os.Getenv("RABBITMQ_URL") != ""
	
	w.Write([]byte(`{
		"status": "online",
		"redisConnected": false,
		"rabbitmqConnected": ` + os.String(hasRabbit) + `,
		"timestamp": "now"
	}`))
}
