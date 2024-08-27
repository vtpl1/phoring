package rtsp

import (
	"fmt"
	"strings"

	"github.com/vtpl1/phoring/backend/sdp"
)

type Media struct {
	Kind      string   `json:"kind,omitempty"`      // video or audio
	Direction string   `json:"direction,omitempty"` // sendonly, recvonly
	Codecs    []*Codec `json:"codecs,omitempty"`

	ID string `json:"id,omitempty"` // MID for WebRTC, Control for RTSP
}

func (m *Media) String() string {
	s := fmt.Sprintf("%s, %s", m.Kind, m.Direction)
	for _, codec := range m.Codecs {
		name := codec.String()

		if strings.Contains(s, name) {
			continue
		}

		s += ", " + name
	}
	return s
}

func (m *Media) Equal(media *Media) bool {
	if media.ID != "" {
		return m.ID == media.ID
	}
	return m.String() == media.String()
}

func MarshalSDP(name string, medias []*Media) ([]byte, error) {
	sd := &sdp.SessionDescription{
		Origin: sdp.Origin{
			Username: "-", SessionID: 1, SessionVersion: 1,
			NetworkType: "IN", AddressType: "IP4", UnicastAddress: "0.0.0.0",
		},
		SessionName: sdp.SessionName(name),
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN", AddressType: "IP4", Address: &sdp.Address{
				Address: "0.0.0.0",
			},
		},
		TimeDescriptions: []sdp.TimeDescription{
			{Timing: sdp.Timing{}},
		},
	}

	for _, media := range medias {
		if media.Codecs == nil {
			continue
		}

		codec := media.Codecs[0]

		switch codec.Name {
		case CodecELD:
			name = CodecAAC
		case CodecPCML:
			name = CodecPCM // beacuse we using pcm.LittleToBig for RTSP server
		default:
			name = codec.Name
		}

		md := &sdp.MediaDescription{
			MediaName: sdp.MediaName{
				Media:  media.Kind,
				Protos: []string{"RTP", "AVP"},
			},
		}
		md.WithCodec(codec.PayloadType, name, codec.ClockRate, codec.Channels, codec.FmtpLine)

		if media.ID != "" {
			md.WithValueAttribute("control", media.ID)
		}

		sd.MediaDescriptions = append(sd.MediaDescriptions, md)
	}

	return sd.Marshal()
}

func UnmarshalMedia(md *sdp.MediaDescription) *Media {
	m := &Media{
		Kind: md.MediaName.Media,
	}

	for _, attr := range md.Attributes {
		switch attr.Key {
		case DirectionSendonly, DirectionRecvonly, DirectionSendRecv:
			m.Direction = attr.Key
		case "control", "mid":
			m.ID = attr.Value
		}
	}

	for _, format := range md.MediaName.Formats {
		m.Codecs = append(m.Codecs, UnmarshalCodec(md, format))
	}

	return m
}
