package main

import (
	"fmt"
	"strings"
	"time"
)

func exportToCsv(runs []TestRun) string {
	csvLines := []string{
		strings.Join([]string{
			"id", "protocol", "enviroment", "time_slot", "test_begin", "test_end", "client_id", "parallel_clients",
			"transfer_start_unix", "transfer_end_unix", "latency_ms", "throughput_mbps", "bytes_sent_total", "bytes_payload", "bandwidth_efficiency",
			"cpu_client_percent_before", "cpu_client_percent_after", "cpu_client_percent_while", "cpu_server_percent_before", "cpu_server_percent_after", "cpu_server_percent_while",
			"ram_client_bytes_before", "ram_client_bytes_after", "ram_client_bytes_while", "ram_server_bytes_before", "ram_server_bytes_after", "ram_server_bytes_while",
			"lost_packets", "retransmissions", "connection_duration", "stream_duration", "error",
		}, ";"),
	}

	for _, run := range runs {
		csvLines = append(csvLines, strings.Join([]string{
			fmt.Sprintf("%d", run.ID), string(run.Protocol), string(run.Enviroment), string(run.TimeSlot), run.TestBegin.Format(time.RFC3339), run.TestEnd.Format(time.RFC3339), fmt.Sprintf("%d", run.ClientID), fmt.Sprintf("%d", run.ParallelClients),
			fmt.Sprintf("%d", run.TransferStartUnix), fmt.Sprintf("%d", run.TransferEndUnix), fmt.Sprintf("%d", run.LatencyMs()), fmt.Sprintf("%f", run.ThroughputMbps), fmt.Sprintf("%d", run.BytesSentTotal), fmt.Sprintf("%d", run.BytesPayload), fmt.Sprintf("%f", run.BandwidthEfficiency()),
			fmt.Sprintf("%f", run.CpuClientPercentBefore), fmt.Sprintf("%f", run.CpuClientPercentAfter), fmt.Sprintf("%f", run.CpuClientPercentWhile), fmt.Sprintf("%f", run.CpuServerPercentBefore), fmt.Sprintf("%f", run.CpuServerPercentAfter), fmt.Sprintf("%f", run.CpuServerPercentWhile),
			fmt.Sprintf("%d", run.RamClientBytesBefore), fmt.Sprintf("%d", run.RamClientBytesAfter), fmt.Sprintf("%d", run.RamClientBytesWhile), fmt.Sprintf("%d", run.RamServerBytesBefore), fmt.Sprintf("%d", run.RamServerBytesAfter), fmt.Sprintf("%d", run.RamServerBytesWhile),
			fmt.Sprintf("%d", run.LostPackets), fmt.Sprintf("%d", run.Retransmissions), fmt.Sprintf("%d", run.ConnectionDuration), fmt.Sprintf("%d", run.StreamDuration), run.Error,
		}, ";"))
	}

	return strings.Join(csvLines, "\n")
}
