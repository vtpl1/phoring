package rtsp

import (
	"bytes"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/vtpl1/phoring/backend/sdp"
)

const (
	DirectionRecvonly = "recvonly"
	DirectionSendonly = "sendonly"
	DirectionSendRecv = "sendrecv"
)

const (
	KindVideo = "video"
	KindAudio = "audio"
)

const (
	CodecH264 = "H264" // payloadType: 96
	CodecH265 = "H265"
	CodecVP8  = "VP8"
	CodecVP9  = "VP9"
	CodecAV1  = "AV1"
	CodecJPEG = "JPEG" // payloadType: 26
	CodecRAW  = "RAW"

	CodecPCMU = "PCMU" // payloadType: 0
	CodecPCMA = "PCMA" // payloadType: 8
	CodecAAC  = "MPEG4-GENERIC"
	CodecOpus = "OPUS" // payloadType: 111
	CodecG722 = "G722"
	CodecMP3  = "MPA" // payload: 14, aka MPEG-1 Layer III
	CodecPCM  = "L16" // Linear PCM (big endian)

	CodecPCML = "PCML" // Linear PCM (little endian)

	CodecELD  = "ELD" // AAC-ELD
	CodecFLAC = "FLAC"

	CodecAll = "ALL"
	CodecAny = "ANY"
)

const sdpHeader = `v=0
o=- 0 0 IN IP4 0.0.0.0
s=-
t=0 0`

func UnmarshalSDP(rawSDP []byte) ([]*Media, error) {
	sd := &sdp.SessionDescription{}
	if err := sd.Unmarshal(rawSDP); err != nil {
		// fix multiple `s=` https://github.com/AlexxIT/WebRTC/issues/417
		re, _ := regexp.Compile("\ns=[^\n]+")
		rawSDP = re.ReplaceAll(rawSDP, nil)

		// fix SDP header for some cameras
		if i := bytes.Index(rawSDP, []byte("\nm=")); i > 0 {
			rawSDP = append([]byte(sdpHeader), rawSDP[i:]...)
		}

		// Fix invalid media type (errSDPInvalidValue) caused by
		// some TP-LINK IP camera, e.g. TL-IPC44GW
		m := regexp.MustCompile("m=[^ ]+ ")
		for _, i := range m.FindAll(rawSDP, -1) {
			switch string(i[2 : len(i)-1]) {
			case "audio", "video", "application":
			default:
				rawSDP = bytes.Replace(rawSDP, i, []byte("m=application "), 1)
			}
		}

		if err == io.EOF {
			rawSDP = append(rawSDP, '\n')
		}

		sd = &sdp.SessionDescription{}
		err = sd.Unmarshal(rawSDP)
		if err != nil {
			return nil, err
		}
	}

	// fix buggy camera https://github.com/AlexxIT/go2rtc/issues/771
	forceDirection := sd.Origin.Username == "CV-RTSPHandler"

	var medias []*Media

	for _, md := range sd.MediaDescriptions {
		media := UnmarshalMedia(md)

		// Check buggy SDP with fmtp for H264 on another track
		// https://github.com/AlexxIT/WebRTC/issues/419
		for _, codec := range media.Codecs {
			if codec.Name == CodecH264 && codec.FmtpLine == "" {
				codec.FmtpLine = findFmtpLine(codec.PayloadType, sd.MediaDescriptions)
			}
		}

		if media.Direction == "" || forceDirection {
			media.Direction = DirectionRecvonly
		}

		medias = append(medias, media)
	}

	return medias, nil
}

func findFmtpLine(payloadType uint8, descriptions []*sdp.MediaDescription) string {
	s := strconv.Itoa(int(payloadType))
	for _, md := range descriptions {
		codec := UnmarshalCodec(md, s)
		if codec.FmtpLine != "" {
			return codec.FmtpLine
		}
	}
	return ""
}

// urlParse fix bugs:
// 1. Content-Base: rtsp://::ffff:192.168.1.123/onvif/profile.1/
// 2. Content-Base: rtsp://rtsp://turret2-cam.lan:554/stream1/
func urlParse(rawURL string) (*url.URL, error) {
	if strings.HasPrefix(rawURL, "rtsp://rtsp://") {
		rawURL = rawURL[7:]
	}

	u, err := url.Parse(rawURL)
	if err != nil && strings.HasSuffix(err.Error(), "after host") {
		if i1 := strings.Index(rawURL, "://"); i1 > 0 {
			if i2 := strings.IndexByte(rawURL[i1+3:], '/'); i2 > 0 {
				return urlParse(rawURL[:i1+3+i2] + ":" + rawURL[i1+3+i2:])
			}
		}
	}

	return u, err
}

func Atoi(s string) (i int) {
	if s != "" {
		i, _ = strconv.Atoi(s)
	}
	return
}
