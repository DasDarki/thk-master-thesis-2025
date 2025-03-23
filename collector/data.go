package main

import (
	"time"

	"github.com/google/uuid"
)

type Protocol string

const (
	ProtocolHTTP3        Protocol = "http3"
	ProtocolWebTransport Protocol = "webtransport"
	ProtocolWebSockets   Protocol = "websockets"
	ProtocolWebRTC       Protocol = "webrtc"
)

type Enviroment string

const (
	EnviromentLocal  Enviroment = "local"
	EnviromentRemote Enviroment = "remote"
)

type TimeSlot string

const (
	TimeSlotMorning   TimeSlot = "morning"
	TimeSlotAfternoon TimeSlot = "afternoon"
	TimeSlotEvening   TimeSlot = "evening"
	TimeSlotNight     TimeSlot = "night"
)

type TestRun struct {
	ID              uuid.UUID
	Protocol        Protocol
	Enviroment      Enviroment
	TimeSlot        TimeSlot
	TestBegin       time.Time
	TestEnd         time.Time
	ClientID        int // used for parallel runs identification
	ParallelClients int // number of parallel clients (used for parallel runs identification)
	Data            TestRunData
}

type TestRunData struct {
	TransferStartUnix      int64   // unix timestamp in milliseconds when the transfer started
	TransferEndUnix        int64   // unix timestamp in milliseconds when the transfer ended
	LatencyMs              int64   // difference between TransferStartUnix and TransferEndUnix
	ThroughputMbps         float64 // throughput in Mbps
	BytesSentTotal         int64   // total bytes sent
	BytesPayload           int64   // bytes sent excluding headers
	BandwidthEfficiency    float64 // BytesPayload / BytesSentTotal
	CpuClientPercentBefore float64 // CPU usage of the client before the transfer
	CpuClientPercentAfter  float64 // CPU usage of the client after the transfer
	CpuClientPercentWhile  float64 // CPU usage of the client while the transfer
	CpuServerPercentBefore float64 // CPU usage of the server before the transfer
	CpuServerPercentAfter  float64 // CPU usage of the server after the transfer
	CpuServerPercentWhile  float64 // CPU usage of the server while the transfer
	RamClientBytesBefore   int64   // RAM usage of the client before the transfer
	RamClientBytesAfter    int64   // RAM usage of the client after the transfer
	RamClientBytesWhile    int64   // RAM usage of the client while the transfer
	RamServerBytesBefore   int64   // RAM usage of the server before the transfer
	RamServerBytesAfter    int64   // RAM usage of the server after the transfer
	RamServerBytesWhile    int64   // RAM usage of the server while the transfer
	LostPackets            int     // number of lost packets
	Retransmissions        int     // number of retransmissions
	ConnectionDuration     int64   // duration of the connection in seconds
	StreamDuration         int64   // duration of the stream in seconds
	Error                  string  // error message if the test failed, empty string otherwise
}
