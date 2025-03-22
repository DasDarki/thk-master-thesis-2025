package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gorilla/websocket"
)

var assetsDir = path.Join("..", "assets")
var videoFile = path.Join(assetsDir, "sample_video.mp4")
var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /")

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello World"))
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/stream", streamVideo)

	if err := http.ListenAndServe("0.0.0.0:2503", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func streamVideo(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /stream")
	chunkSize := 16384

	q := r.URL.Query()
	if q.Get("chunkSize") != "" {
		chunkSize2, err := strconv.Atoi(q.Get("chunkSize"))
		if err != nil {
			log.Printf("Failed to parse chunkSize: %v", err)
			http.Error(w, "Failed to parse chunkSize", http.StatusBadRequest)
			return
		}

		chunkSize = chunkSize2
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	log.Printf("Upgraded to WebSocket!")

	defer conn.Close()

	file, err := os.Open(videoFile)
	if err != nil {
		log.Printf("Failed to open video file: %v", err)
		http.Error(w, "Failed to open video file", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Printf("Failed to read video file: %v", err)
			http.Error(w, "Failed to read video file", http.StatusInternalServerError)
			return
		}

		if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
			log.Printf("Failed to write message: %v", err)
			http.Error(w, "Failed to write message", http.StatusInternalServerError)
			return
		}
	}

	log.Println("Video sent")
}
