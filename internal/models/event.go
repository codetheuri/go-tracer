package models

import "time"

// FlowEvent represents the summarized data from a completed network flow.
type FlowEvent struct {
	FlowID    string    `json:"flow_id"`
	Timestamp time.Time `json:"timestamp"` 
	Interface string    `json:"interface"`
	Hostname  string    `json:"hostname"`
	Direction string    `json:"direction"`

	// 5-tuple
	SrcIP    string `json:"src_ip"`
	DstIP    string `json:"dst_ip"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Protocol string `json:"protocol"`

	FlowDuration  float64 `json:"flow_duration"`  // in seconds
	PacketCount   int64   `json:"packet_count"`   // Total packets in flow
	ByteCount     int64   `json:"byte_count"`     // Total bytes in flow
	AvgPacketSize float64 `json:"avg_packet_size"`

}