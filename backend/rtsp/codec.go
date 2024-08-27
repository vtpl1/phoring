package rtsp

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/vtpl1/phoring/backend/sdp"
)

type Codec struct {
	Name        string // H264, PCMU, PCMA, opus...
	ClockRate   uint32 // 90000, 8000, 16000...
	Channels    uint16 // 0, 1, 2
	FmtpLine    string
	PayloadType uint8
}

func (c *Codec) String() (s string) {
	s = c.Name
	if c.ClockRate != 0 && c.ClockRate != 90000 {
		s += fmt.Sprintf("/%d", c.ClockRate)
	}
	if c.Channels > 0 {
		s += fmt.Sprintf("/%d", c.Channels)
	}
	return
}

func UnmarshalCodec(md *sdp.MediaDescription, payloadType string) *Codec {
	c := &Codec{PayloadType: byte(Atoi(payloadType))}

	for _, attr := range md.Attributes {
		switch {
		case c.Name == "" && attr.Key == "rtpmap" && strings.HasPrefix(attr.Value, payloadType):
			i := strings.IndexByte(attr.Value, ' ')
			ss := strings.Split(attr.Value[i+1:], "/")

			c.Name = strings.ToUpper(ss[0])
			// fix tailing space: `a=rtpmap:96 H264/90000 `
			c.ClockRate = uint32(Atoi(strings.TrimRightFunc(ss[1], unicode.IsSpace)))

			if len(ss) == 3 && ss[2] == "2" {
				c.Channels = 2
			}
		case c.FmtpLine == "" && attr.Key == "fmtp" && strings.HasPrefix(attr.Value, payloadType):
			if i := strings.IndexByte(attr.Value, ' '); i > 0 {
				c.FmtpLine = attr.Value[i+1:]
			}
		}
	}

	switch c.Name {
	case "PCM":
		// https://www.reddit.com/r/Hikvision/comments/17elxex/comment/k642g2r/
		// check pkg/rtsp/rtsp_test.go TestHikvisionPCM
		c.Name = CodecPCML
	case "":
		// https://en.wikipedia.org/wiki/RTP_payload_formats
		switch payloadType {
		case "0":
			c.Name = CodecPCMU
			c.ClockRate = 8000
		case "8":
			c.Name = CodecPCMA
			c.ClockRate = 8000
		case "10":
			c.Name = CodecPCM
			c.ClockRate = 44100
			c.Channels = 2
		case "11":
			c.Name = CodecPCM
			c.ClockRate = 44100
		case "14":
			c.Name = CodecMP3
			c.ClockRate = 90000 // it's not real sample rate
		case "26":
			c.Name = CodecJPEG
			c.ClockRate = 90000
		case "96", "97", "98":
			if len(md.Bandwidth) == 0 {
				c.Name = payloadType
				break
			}

			// FFmpeg + RTSP + pcm_s16le = doesn't pass info about codec name and params
			// so try to guess the codec based on bitrate
			// https://github.com/AlexxIT/go2rtc/issues/523
			switch md.Bandwidth[0].Bandwidth {
			case 128:
				c.ClockRate = 8000
			case 256:
				c.ClockRate = 16000
			case 384:
				c.ClockRate = 24000
			case 512:
				c.ClockRate = 32000
			case 705:
				c.ClockRate = 44100
			case 768:
				c.ClockRate = 48000
			case 1411:
				// default Windows DShow
				c.ClockRate = 44100
				c.Channels = 2
			case 1536:
				// default Linux ALSA
				c.ClockRate = 48000
				c.Channels = 2
			default:
				c.Name = payloadType
				break
			}

			c.Name = CodecPCML
		default:
			c.Name = payloadType
		}
	}

	return c
}
