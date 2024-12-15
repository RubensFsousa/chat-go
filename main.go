package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/coder/websocket"
)

type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

var (
	clients   = make(map[*websocket.Conn]string)
	clientsMu sync.RWMutex
)

func broadcast(ctx context.Context, sender *websocket.Conn, message Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error serializing message: %v", err)
		return
	}

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	for client := range clients {
		if client == sender {
			continue
		}

		if err := client.Write(ctx, websocket.MessageText, messageJSON); err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close(websocket.StatusInternalError, "Error sending message")
			removeClient(client)
		}
	}
}

func removeClient(client *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, client)
}

func addClient(client *websocket.Conn, username string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[client] = username
}

func counterHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(strconv.Itoa(len(clients)))); err != nil {
		log.Printf("Error sending message to client: %v", err)
	}
}

func webHandler(w http.ResponseWriter, r *http.Request) {
	con, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("Error accepting connection: %v", err)
		return
	}
	defer con.Close(websocket.StatusNormalClosure, "")

	_, usernameBytes, err := con.Read(r.Context())
	if err != nil {
		log.Printf("Error reading username: %v", err)
		return
	}

	username := string(usernameBytes)
	addClient(con, username)
	log.Printf("New client connected: %s", username)

	for {
		_, messageBytes, err := con.Read(r.Context())
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			removeClient(con)
			return
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		message.Username = username
		broadcast(r.Context(), con, message)
	}
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", webHandler)
	http.HandleFunc("/clients-count", counterHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
