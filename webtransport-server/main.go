package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
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

	runID, err := strconv.Atoi(r.URL.Query().Get("runID"))
	if err != nil {
		http.Error(w, "Invalid runID", http.StatusBadRequest)
		return
	}

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

	transferStart := time.Now().Unix()
	cpuBefore := getCpuUsagePercentage()
	ramBefore := getRamUsageBytes()
	cpuWhile := float64(0)
	ramWhile := uint64(0)

	go func() {
		time.Sleep(500 * time.Millisecond)
		cpuWhile = getCpuUsagePercentage()
		ramWhile = getRamUsageBytes()
	}()

	_, err = io.Copy(stream, file)
	if err != nil {
		log.Printf("Error while streaming: %v", err)
		return
	}

	stream.Close()

	cpuAfter := getCpuUsagePercentage()
	ramAfter := getRamUsageBytes()

	collectMetrics(runID, map[string]any{
		"TransferStartUnix":      transferStart,
		"BytesPayload":           stat.Size(),
		"CpuServerPercentBefore": cpuBefore,
		"CpuServerPercentWhile":  cpuWhile,
		"CpuServerPercentAfter":  cpuAfter,
		"RamServerBytesBefore":   ramBefore,
		"RamServerBytesWhile":    ramWhile,
		"RamServerBytesAfter":    ramAfter,
	})

	log.Println("Streaming finished successfully")
}

func getCpuUsagePercentage() float64 {
	percentages, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil || len(percentages) == 0 {
		return 0.0
	}
	return percentages[0]
}

func getRamUsageBytes() uint64 {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return vmStat.Used
}
