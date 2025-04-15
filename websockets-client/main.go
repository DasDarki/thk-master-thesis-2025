package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const URL = "ws://localhost:2503"
const REMOTE_URL = "wss://thkm25_websockets.nauri.io"

func main() {
	url, runID := parseArguments()

	conn, _, err := websocket.DefaultDialer.Dial(url+"/stream?runID="+fmt.Sprintf("%d", runID), nil)
	if err != nil {
		collectMetrics(runID, map[string]any{
			"@end":  true,
			"Error": fmt.Sprintf("Failed to dial WebSockets: %v", err),
		})
		log.Fatalf("Failed to dial WebSockets: %v", err)
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

	defer conn.Close()

	file, err := os.Create(fmt.Sprintf("output_%d.mp4", runID))
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}

	defer file.Close()

	defer func() {
		cpuAfter := getCpuUsagePercentage()
		ramAfter := getRamUsageBytes()
		lostAfter, recvAfter := getPacketStats()

		collectMetrics(runID, map[string]any{
			"@end":                   true,
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

		log.Printf("Connection duration: %d ms", time.Since(connectEstablishTime).Milliseconds())
		log.Printf("CPU usage before: %.2f%%", cpuBefore)
		log.Printf("CPU usage while: %.2f%%", cpuWhile)
		log.Printf("CPU usage after: %.2f%%", cpuAfter)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
				break
			}

			collectMetrics(runID, map[string]any{
				"@end":  true,
				"Error": fmt.Sprintf("Failed to read message: %v", err),
			})

			log.Fatalf("Failed to read message: %v", err)
		}

		n, err := file.Write(message)
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}

		log.Printf("Wrote %d bytes", n)
	}

	log.Println("Done")
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
