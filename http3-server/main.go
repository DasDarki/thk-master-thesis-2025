package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/quic-go/quic-go/http3"
)

var assetsDir = path.Join("..", "assets")
var videoFile = path.Join(assetsDir, "sample_video.mp4")

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /")

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello World"))
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/stream", streamVideo)

	if err := http3.ListenAndServeQUIC("0.0.0.0:2501", path.Join(assetsDir, "ssl_localhost.crt"), path.Join(assetsDir, "ssl_localhost.key"), mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func streamVideo(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /stream")

	video, err := os.Open(videoFile)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	defer video.Close()

	stat, err := video.Stat()
	if err != nil {
		http.Error(w, "Could not obtain file info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, videoFile, stat.ModTime(), video)
}
