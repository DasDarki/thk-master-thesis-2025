package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

var assetsDir = path.Join("..", "assets")
var videoFile = path.Join(assetsDir, "sample_video.mp4")

func main() {
	stat, err := os.Stat(videoFile)
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}

	log.Printf("File %s exists, size: %d bytes", videoFile, stat.Size())

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

	runID, err := strconv.Atoi(r.URL.Query().Get("runID"))
	if err != nil {
		http.Error(w, "Invalid runID", http.StatusBadRequest)
		return
	}

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

	http.ServeContent(w, r, videoFile, stat.ModTime(), video)

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
