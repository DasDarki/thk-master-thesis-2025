package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func saveToCsv(run *TestRun) error {
	return writeToCsv("results.csv", run)
}

func saveToEmergencyCsv(run *TestRun) error {
	return writeToCsv("emergency.csv", run)
}

func writeToCsv(file string, run *TestRun) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.Println("Creating new CSV file", file)
		parts := []string{
			"id", "protocol", "enviroment", "time_slot", "test_begin", "test_end", "client_id", "parallel_clients",
			"transfer_start_unix", "transfer_end_unix", "latency_ms", "throughput_mbps", "bytes_sent_total", "bytes_payload", "bandwidth_efficiency",
			"cpu_client_percent_before", "cpu_client_percent_after", "cpu_client_percent_while", "cpu_server_percent_before", "cpu_server_percent_after", "cpu_server_percent_while",
			"ram_client_bytes_before", "ram_client_bytes_after", "ram_client_bytes_while", "ram_server_bytes_before", "ram_server_bytes_after", "ram_server_bytes_while",
			"lost_packets", "retransmissions", "connection_duration", "stream_duration", "error",
		}

		if err := os.WriteFile(file, []byte(strings.Join(parts, ";")+"\n"), 0644); err != nil {
			return err
		}
	}

	data := []string{
		run.ID.String(), string(run.Protocol), string(run.Enviroment), string(run.TimeSlot), run.TestBegin.Format(time.RFC3339), run.TestEnd.Format(time.RFC3339), fmt.Sprintf("%d", run.ClientID), fmt.Sprintf("%d", run.ParallelClients),
		fmt.Sprintf("%d", run.Data.TransferStartUnix), fmt.Sprintf("%d", run.Data.TransferEndUnix), fmt.Sprintf("%d", run.Data.LatencyMs), fmt.Sprintf("%f", run.Data.ThroughputMbps), fmt.Sprintf("%d", run.Data.BytesSentTotal), fmt.Sprintf("%d", run.Data.BytesPayload), fmt.Sprintf("%f", run.Data.BandwidthEfficiency),
		fmt.Sprintf("%f", run.Data.CpuClientPercentBefore), fmt.Sprintf("%f", run.Data.CpuClientPercentAfter), fmt.Sprintf("%f", run.Data.CpuClientPercentWhile), fmt.Sprintf("%f", run.Data.CpuServerPercentBefore), fmt.Sprintf("%f", run.Data.CpuServerPercentAfter), fmt.Sprintf("%f", run.Data.CpuServerPercentWhile),
		fmt.Sprintf("%d", run.Data.RamClientBytesBefore), fmt.Sprintf("%d", run.Data.RamClientBytesAfter), fmt.Sprintf("%d", run.Data.RamClientBytesWhile), fmt.Sprintf("%d", run.Data.RamServerBytesBefore), fmt.Sprintf("%d", run.Data.RamServerBytesAfter), fmt.Sprintf("%d", run.Data.RamServerBytesWhile),
		fmt.Sprintf("%d", run.Data.LostPackets), fmt.Sprintf("%d", run.Data.Retransmissions), fmt.Sprintf("%d", run.Data.ConnectionDuration), fmt.Sprintf("%d", run.Data.StreamDuration), run.Data.Error,
	}

	return appendToEndOfFile(file, strings.Join(data, ";")+"\n")
}

func appendToEndOfFile(file string, data string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(data); err != nil {
		return err
	}

	return nil
}
