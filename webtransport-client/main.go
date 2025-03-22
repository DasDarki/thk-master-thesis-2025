package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"

	"github.com/quic-go/webtransport-go"
)

const URL = "https://localhost:2504"

func main() {
	d := webtransport.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	resp, sess, err := d.Dial(context.Background(), URL+"/stream", nil)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	defer sess.CloseWithError(0, "bye")

	log.Printf("Connected: %v", resp.Status)

	stream, err := sess.AcceptUniStream(context.Background())
	if err != nil {
		log.Fatalf("Could not accept stream: %v", err)
	}

	file, err := os.Create("output.mp4")
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}

	defer file.Close()

	n, err := io.Copy(file, stream)
	if err != nil {
		log.Fatalf("Could not copy stream data: %v", err)
	}

	log.Printf("Successfully received %d bytes", n)
}
