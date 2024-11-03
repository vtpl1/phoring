package monitor

import "github.com/google/uuid"

type Device struct {
	ID   uuid.UUID `json:"id,omitempty"`
	IP   string    `json:"ip,omitempty"`
	Name string    `json:"name,omitempty"`
}

func NewDevice() Device {
	return Device{}
}
