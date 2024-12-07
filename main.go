package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func webHandler(w http.ResponseWriter, r *http.Request) {
	con, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"www.google.com"},
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, data, err := con.Read(r.Context())
		if err != nil {
			log.Fatal(err)
			break
		}

		log.Println(string(data))

		con.Write(r.Context(), websocket.MessageText, []byte("pong"))
	}

}

func main() {

	http.HandleFunc("/", webHandler)

	http.ListenAndServe(":8080", nil)
}
