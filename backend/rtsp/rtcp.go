package rtsp

import "github.com/vtpl1/phoring/backend/rtcp"

type RTCP struct {
	Channel byte
	Header  rtcp.Header
	Packets []rtcp.Packet
}
