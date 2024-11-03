package monitor

import (
	"reflect"
	"testing"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		name string
		want Device
	}{
		{
			name: "empty",
			want: Device{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDevice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}
