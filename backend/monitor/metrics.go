package monitor

import "time"

type Metrics struct {
	ID               int64     `json:"id,omitempty"`
	Timestamp        time.Time `json:"timestamp,omitempty"`
	CPUUsage         []float64 `json:"cpu_usage,omitempty"`
	MemoryUsed       uint64    `json:"memory_used,omitempty"`
	MemoryTotal      uint64    `json:"memory_total,omitempty"`
	NetworkBytesRecv uint64    `json:"network_bytes_recv,omitempty"`
	NetworkBytesSent uint64    `json:"network_bytes_sent,omitempty"`
}
