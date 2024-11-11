package monitor

import (
	"errors"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/vtpl1/phoring/backend/cmd/utils"
)

type Metrics struct {
	ID               int64   `json:"id,omitempty"`
	Name             string  `json:"name,omitempty"`
	Timestamp        int64   `json:"timestamp,omitempty"`
	CPUPercent       float64 `json:"cpuPercent,omitempty"`
	MemoryPercent    float32 `json:"memoryPercent,omitempty"`
	MemoryTotal      uint64  `json:"memoryTotal,omitempty"`
	ThreadCount      int32   `json:"threadCount,omitempty"`
	OpenFiles        int32   `json:"openFiles,omitempty"`
	NetworkBytesRecv uint64  `json:"networkBytesRecv,omitempty"`
	NetworkBytesSent uint64  `json:"networkBytesSent,omitempty"`
}

func GetMetrics(processName string) (Metrics, error) {
	log := utils.GetLogger("")
	timeStamp := time.Now().UnixMilli()
	metrics := Metrics{Name: processName, Timestamp: timeStamp}
	processes, err := process.Processes()
	if err != nil {
		log.Error().Str("process", processName).Err(err).Msgf("Error getting processes")
		return Metrics{}, err
	}
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		if strings.ToLower(name) == processName {
			cpuPercent, err := p.CPUPercent()
			if err != nil {
				log.Error().Str("process", processName).Err(err).Send()
			}
			memPercent, err := p.MemoryPercent()
			if err != nil {
				log.Error().Str("process", processName).Err(err).Send()
			}
			threadCount, err := p.NumThreads()
			if err != nil {
				log.Error().Str("process", processName).Err(err).Send()
			}
			openFiles, err := p.OpenFiles()
			if err != nil {
				log.Error().Str("process", processName).Err(err).Send()
			}
			metrics.CPUPercent = cpuPercent
			metrics.MemoryPercent = memPercent
			metrics.ThreadCount = threadCount
			metrics.OpenFiles = int32(len(openFiles))
			return metrics, nil
		}
	}

	return metrics, errors.New("Process not found")
}
