package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

var assetsDir = path.Join("..", "assets")
var videoFile = path.Join(assetsDir, "sample_video.mp4")
var webtransportSrv *webtransport.Server

func main() {
	webtransportSrv = &webtransport.Server{
		H3: http3.Server{
			Addr: "0.0.0.0:2504",
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /")

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello World"))
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/stream", streamVideo)

	if err := webtransportSrv.ListenAndServeTLS(path.Join(assetsDir, "ssl_localhost.crt"), path.Join(assetsDir, "ssl_localhost.key")); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func streamVideo(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /stream")

	sess, err := webtransportSrv.Upgrade(w, r)
	if err != nil {
		log.Printf("Failed to upgrade to WebTransport: %v", err)
		http.Error(w, "Failed to upgrade to WebTransport", http.StatusInternalServerError)
		return
	}

	log.Printf("Upgraded to WebTransport: %v", sess)

	stream, err := sess.OpenUniStream()
	if err != nil {
		log.Printf("Failed to open stream: %v", err)
		http.Error(w, "Failed to open stream", http.StatusInternalServerError)
		return
	}

	log.Printf("Opened stream: %v", stream)

	file, err := os.Open(videoFile)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not obtain file info", http.StatusInternalServerError)
		return
	}

	log.Printf("Streaming video (%d bytes): %s", stat.Size(), videoFile)

	_, err = io.Copy(stream, file)
	if err != nil {
		log.Printf("Error while streaming: %v", err)
		return
	}

	stream.Close()
	log.Println("Streaming finished successfully")
}
