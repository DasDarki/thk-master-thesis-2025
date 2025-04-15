package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const URL = "https://localhost:2501"
const REMOTE_URL = "https://thkm25_http3.nauri.io:2501"

func main() {
	url, runID := parseArguments()

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

	connectEstablishTime := time.Now()
	cpuBefore := getCpuUsagePercentage()
	ramBefore := getRamUsageBytes()
	cpuWhile := float64(0)
	ramWhile := uint64(0)
	lost, recv := getPacketStats()

	go func() {
		time.Sleep(500 * time.Millisecond)
		cpuWhile = getCpuUsagePercentage()
		ramWhile = getRamUsageBytes()
	}()

	resp, err := client.Get(url + "/stream?runID=" + fmt.Sprintf("%d", runID))
	if err != nil {
		collectMetrics(runID, map[string]any{
			"error": fmt.Sprintf("Failed to GET: %v", err),
		})
		log.Fatalf("Failed to GET: %v", err)
	}

	defer resp.Body.Close()

	output, err := os.Create(fmt.Sprintf("output_%d.mp4", runID))
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	defer output.Close()

	for {
		buf := make([]byte, 1024)
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == http.ErrBodyReadAfterClose || err == io.EOF || err.Error() == "204 No Content" {
				log.Println("Connection closed by server")
			} else {
				collectMetrics(runID, map[string]any{
					"error": fmt.Sprintf("Failed to read response body: %v", err),
				})
				log.Fatalf("Failed to read response body: %v", err)
			}
			break
		}

		output.Write(buf[:n])
	}

	cpuAfter := getCpuUsagePercentage()
	ramAfter := getRamUsageBytes()
	lostAfter, recvAfter := getPacketStats()

	collectMetrics(runID, map[string]any{
		"TransferEndUnix":        time.Now().Unix(),
		"ConnectionDuration":     time.Since(connectEstablishTime).Milliseconds(),
		"CpuClientPercentBefore": cpuBefore,
		"CpuClientPercentWhile":  cpuWhile,
		"CpuClientPercentAfter":  cpuAfter,
		"RamClientBytesBefore":   ramBefore,
		"RamClientBytesWhile":    ramWhile,
		"RamClientBytesAfter":    ramAfter,
		"LostPackets":            lostAfter - lost,
		"BytesSentTotal":         recvAfter - recv,
	})
}

func parseArguments() (string, int) {
	a := os.Args[1:]
	isLocal := false
	runId := -1

	for _, arg := range a {
		if arg == "-l" {
			isLocal = true
		} else if strings.HasPrefix(arg, "-r") {
			runIDStr := strings.TrimPrefix(arg, "-r")
			runID, err := strconv.Atoi(runIDStr)
			if err != nil {
				log.Fatalf("Invalid runID: %v", err)
			}

			runId = runID
		}
	}

	url := REMOTE_URL
	if isLocal {
		url = URL
	}

	return url, runId
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

// Returns netstat -e output for lost packets, and bytes received (sent)
func getPacketStats() (int64, int64) {
	cmd, err := exec.Command("netstat", "-e").Output()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return 0, 0
	}

	output := string(cmd)
	lines := strings.Split(output, "\n")
	lostPackets := int64(0)
	bytesReceived := int64(0)

	for _, line := range lines {
		if strings.Contains(line, "Bytes") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				r, _ := strconv.ParseInt(parts[1], 10, 64)
				bytesReceived = r
			}
		} else if strings.Contains(line, "Verworfen") || strings.Contains(line, "Lost") || strings.Contains(line, "Fehler") || strings.Contains(line, "Errors") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				r, _ := strconv.ParseInt(parts[1], 10, 64)
				lostPackets += r
			}
		}
	}

	return lostPackets, bytesReceived
}
