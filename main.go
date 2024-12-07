package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/coder/websocket"
)

var (
	clients map[*websocket.Conn]bool = make(map[*websocket.Conn]bool)
)

func webHandler(w http.ResponseWriter, r *http.Request) {
	con, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"www.google.com"},
	})
	if err != nil {
		log.Fatal(err)
	}

	clients[con] = true

	for {
		_, data, err := con.Read(r.Context())
		if err != nil {
			log.Fatal(err)
			delete(clients, con)
			break
		}

		log.Println(string(data))

		message := string(data)

		for client := range clients {
			client.Write(r.Context(), websocket.MessageText, []byte(message))
		}

	}

}

func main() {

	http.HandleFunc("/ws", webHandler)

	http.HandleFunc("/clients-count", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strconv.Itoa(len(clients))))
	})

	http.ListenAndServe(":8080", nil)
}
