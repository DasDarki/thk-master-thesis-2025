package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var assetsDir = path.Join("..", "assets")
var videoFile = path.Join(assetsDir, "sample_video.mp4")
var upgrader = websocket.Upgrader{}

type signal struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

func main() {
	http.HandleFunc("/stream", handleStream)
	log.Println("[server] Listening on :2502")
	log.Fatal(http.ListenAndServe("0.0.0.0:2502", nil))
}

func handleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[server] WebSocket upgrade error:", err)
		return
	}

	log.Println("[server] Client connected via WebSocket")

	go handleConnection(conn)
}

func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Println("[server] Failed to create PeerConnection:", err)
		return
	}

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[server] Connection state: %s", state.String())
		if state == webrtc.PeerConnectionStateConnected {
			log.Println("[server] PeerConnection connected!")
		}
	})

	defer pc.Close()

	log.Println("[server] Created PeerConnection")

	dc, err := pc.CreateDataChannel("video", nil)
	if err != nil {
		log.Println("[server] Failed to create DataChannel:", err)
		return
	}

	dc.OnOpen(func() {
		log.Println("[server] DataChannel open: sending video...")

		file, err := os.Open(videoFile)
		if err != nil {
			log.Println("[server] Failed to open video file:", err)
			return
		}

		defer file.Close()

		buf := make([]byte, 16384)
		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println("[server] Error reading file:", err)
				return
			}
			dc.Send(buf[:n])
		}

		log.Println("[server] File sent. Closing DataChannel.")
		dc.SendText("done")
		dc.Close()
		pc.Close()
		conn.Close()
	})

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Println("[server] Failed to create offer:", err)
		return
	}

	pc.SetLocalDescription(offer)
	<-webrtc.GatheringCompletePromise(pc)

	// SEND OFFER
	offerMsg := signal{Type: "offer", SDP: offer.SDP}
	conn.WriteJSON(offerMsg)
	log.Println("[server] Sent offer")

	// WAIT FOR ANSWER
	var answerMsg signal
	if err := conn.ReadJSON(&answerMsg); err != nil {
		log.Println("[server] Failed to read answer:", err)
		return
	}

	log.Println("[server] Received answer", answerMsg.Type, answerMsg.SDP)

	if answerMsg.Type != "answer" {
		log.Println("[server] Unexpected message:", answerMsg.Type)
		return
	}

	pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  answerMsg.SDP,
	})

	log.Println("[server] Set remote description")

	for {
		// wait for client to close
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Println("[server] Client disconnected:", err)
			break
		}
	}
}
