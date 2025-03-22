package main

import (
	"log"
	"os"

	"github.com/gorilla/websocket"
)

const URL = "ws://localhost:2503"

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(URL+"/stream", nil)
	if err != nil {
		log.Fatalf("Failed to dial WebSockets: %v", err)
	}

	defer conn.Close()

	file, err := os.Create("output.mp4")
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}

	defer file.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}

			log.Fatalf("Failed to read message: %v", err)
		}

		n, err := file.Write(message)
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}

		log.Printf("Wrote %d bytes", n)
	}

	log.Println("Done")
}
