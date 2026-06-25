package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"os"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub хранит активные соединения
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil { log.Fatal(err) }
	defer ws.Close()
	clients[ws] = true
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil { delete(clients, ws); break }
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil { client.Close(); delete(clients, client) }
		}
	}
}

func main() {
	// Берем порт из переменной окружения Render или используем 10000
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./frontend")))
	mux.Handle("/dist/", http.StripPrefix("/dist/", http.FileServer(http.Dir("./frontend/dist"))))
	mux.HandleFunc("/ws", handleConnections)
	
	go handleMessages()

	log.Printf("Сервер запущен на порту %s", port)
	// Слушаем на "0.0.0.0" для всех входящих соединений Render
	err := http.ListenAndServe("0.0.0.0:"+port, mux)
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
