package main

import (
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

const URL = "ws://127.0.0.1:2502"

type signal struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(URL+"/stream", nil)
	if err != nil {
		log.Fatal("[client] WebSocket dial error:", err)
	}

	defer conn.Close()

	var offerMsg signal
	if err := conn.ReadJSON(&offerMsg); err != nil {
		log.Fatal("[client] Failed to read offer:", err)
	}

	if offerMsg.Type != "offer" {
		log.Fatal("[client] Expected offer, got:", offerMsg.Type)
	}

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Fatal("[client] PeerConnection error:", err)
	}

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("[client] Connection state: %s", state.String())
		if state == webrtc.PeerConnectionStateConnected {
			log.Println("[client] PeerConnection connected!")
		} else if state == webrtc.PeerConnectionStateFailed {
			log.Fatal("[client] PeerConnection failed!")
		}
	})

	defer pc.Close()

	file, err := os.Create("output.mp4")
	if err != nil {
		log.Fatal("[client] Failed to create output file:", err)
	}

	defer file.Close()

	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Println("[client] DataChannel received:", dc.Label())

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if msg.IsString && string(msg.Data) == "done" {
				log.Println("[client] Transfer complete. Closing.")
				pc.Close()
				conn.Close()
				os.Exit(0)
			}
			file.Write(msg.Data)
		})
	})

	pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerMsg.SDP,
	})

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Fatal("[client] CreateAnswer failed:", err)
	}

	pc.SetLocalDescription(answer)
	<-webrtc.GatheringCompletePromise(pc)

	answerMsg := signal{Type: "answer", SDP: answer.SDP}
	conn.WriteJSON(answerMsg)

	log.Println("[client] WebRTC setup complete. Receiving file...")
	select {}
}
