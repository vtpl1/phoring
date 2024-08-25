package rtsp

import (
	"net/http"
	"reflect"
	"testing"
)

func TestClient_Options(t *testing.T) {
	for _, tt := range []struct {
		name     string
		c        *Client
		url      string
		wantResp *Response
		wantErr  bool
	}{
		{
			name: "Default options",
			c:    &Client{},
			url:  "",
			wantResp: &Response{
				Proto:      "RTSP",
				ProtoMajor: 1,
				ProtoMinor: 1,
				StatusCode: 200,
				Status:     "OK",
				Header: http.Header{
					"Cseq":   {"1"},
					"Public": {"DESCRIBE, PAUSE, PLAY, SETUP, TEARDOWN"},
				},
			},
			wantErr: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := tt.c.Options(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Options() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResp.String() != tt.wantResp.String() {
				t.Errorf("Client.Options() = [%v], want [%v]", gotResp, tt.wantResp)
			}
		})
	}
}

func TestClient_Describe(t *testing.T) {
	type args struct {
		urlStr string
	}
	tests := []struct {
		name    string
		c       *Client
		args    args
		want    *Response
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Describe(tt.args.urlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Describe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Describe() = %v, want %v", got, tt.want)
			}
		})
	}
}
