package rtsp

import "net/http"

func Options() (resp *http.Response, err error) {
	return http.Get("ddd")
}
