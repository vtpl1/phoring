package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
	ID               int64     `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	CPUUsage         []float64 `json:"cpu_usage"`
	MemoryUsed       uint64    `json:"memory_used"`
	MemoryTotal      uint64    `json:"memory_total"`
	NetworkBytesRecv uint64    `json:"network_bytes_recv"`
	NetworkBytesSent uint64    `json:"network_bytes_sent"`
}

var db *sql.DB

func initDB() {
	var err error
	// Connect to SQLite database
	db, err = sql.Open("sqlite", "./metrics.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			cpu_usage TEXT,
			memory_used INTEGER,
			memory_total INTEGER,
			network_bytes_recv INTEGER,
			network_bytes_sent INTEGER
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func insertMetrics(metrics Metrics) error {
	// Convert CPUUsage array to JSON
	cpuUsageJSON, err := json.Marshal(metrics.CPUUsage)
	if err != nil {
		return err
	}

	// Insert the metrics data into the database
	_, err = db.Exec(`
		INSERT INTO metrics (cpu_usage, memory_used, memory_total, network_bytes_recv, network_bytes_sent)
		VALUES (?, ?, ?, ?, ?)`,
		cpuUsageJSON, metrics.MemoryUsed, metrics.MemoryTotal, metrics.NetworkBytesRecv, metrics.NetworkBytesSent,
	)
	return err
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Collect CPU, memory, and network metrics
	cpuUsage, _ := cpu.Percent(0, false)
	memoryUsage, _ := mem.VirtualMemory()
	networkUsage, _ := net.IOCounters(false)

	// Create a metrics object
	metrics := Metrics{
		CPUUsage:         cpuUsage,
		MemoryUsed:       memoryUsage.Used,
		MemoryTotal:      memoryUsage.Total,
		NetworkBytesRecv: networkUsage[0].BytesRecv,
		NetworkBytesSent: networkUsage[0].BytesSent,
	}

	// Insert metrics data into the database
	err := insertMetrics(metrics)
	if err != nil {
		http.Error(w, "Failed to save metrics to database", http.StatusInternalServerError)
		return
	}

	// Respond with the current metrics as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	// Query the historical metrics from the database
	rows, err := db.Query("SELECT id, timestamp, cpu_usage, memory_used, memory_total, network_bytes_recv, network_bytes_sent FROM metrics ORDER BY timestamp DESC LIMIT 100")
	if err != nil {
		http.Error(w, "Failed to retrieve historical data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var metricsHistory []Metrics

	for rows.Next() {
		var metrics Metrics
		var cpuUsageJSON string

		err := rows.Scan(&metrics.ID, &metrics.Timestamp, &cpuUsageJSON, &metrics.MemoryUsed, &metrics.MemoryTotal, &metrics.NetworkBytesRecv, &metrics.NetworkBytesSent)
		if err != nil {
			http.Error(w, "Error parsing data", http.StatusInternalServerError)
			return
		}

		// Decode the CPU usage from JSON
		json.Unmarshal([]byte(cpuUsageJSON), &metrics.CPUUsage)
		metricsHistory = append(metricsHistory, metrics)
	}

	// Respond with the metrics history as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricsHistory)
}

func main1() {
	// Initialize the SQLite database
	initDB()

	// Serve the current metrics and historical data
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/metrics/history", historyHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {

}
