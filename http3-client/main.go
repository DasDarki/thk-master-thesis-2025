package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const URL = "https://localhost:2501"

func main() {
	tr := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QUICConfig: &quic.Config{},
	}
	defer tr.Close()
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get(URL + "/stream")
	if err != nil {
		log.Fatalf("Failed to GET: %v", err)
	}

	defer resp.Body.Close()

	output, err := os.Create("output.mp4")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	defer output.Close()

	for {
		buf := make([]byte, 1024)
		n, err := resp.Body.Read(buf)
		if err != nil {
			break
		}

		output.Write(buf[:n])
	}
}
