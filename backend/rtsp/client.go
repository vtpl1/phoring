package rtsp

import (
	"bufio"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vtpl1/phoring/backend/rtcp"
	"github.com/vtpl1/phoring/backend/rtp"
)

type Client struct {
	URL         *url.URL
	Backchannel bool
	UserAgent   string
	SDP         string
	Media       string
	Medias      []*Media
	SessionName string
	Timeout     int

	sequence  int
	auth      *Auth
	conn      net.Conn
	session   string
	keepalive int
	uri       string
	timeout   time.Duration
	reader    *bufio.Reader
	mode      Mode
	state     State
	playOK    bool
}

type State byte

func (s State) String() string {
	switch s {
	case StateNone:
		return "NONE"
	case StateConn:

		return "CONN"
	case StateSetup:
		return SETUP
	case StatePlay:
		return PLAY
	}
	return strconv.Itoa(int(s))
}

const (
	StateNone State = iota
	StateConn
	StateSetup
	StatePlay
)

var (
	errNotImplemented = errors.New("not implemented")
)

const (
	BufferSize      = 64 * 1024 // 64K
	ConnDialTimeout = time.Second * 3
	ConnDeadline    = time.Second * 5
	ProbeTimeout    = time.Second * 3
)

type Mode byte

const (
	ModeActiveProducer Mode = iota + 1 // typical source (client)
	ModePassiveConsumer
	ModePassiveProducer
	ModeActiveConsumer
)

const EndLine = "\r\n"

const (
	// Protocol RTSP version 1.0
	ProtoRTSP = "RTSP/1.0"
	// Client to server for presentation and stream objects; recommended
	DESCRIBE = "DESCRIBE"
	// Bidirectional for client and stream objects; optional
	ANNOUNCE = "ANNOUNCE"
	// Bidirectional for client and stream objects; optional
	GET_PARAMETER = "GET_PARAMETER"
	// Bidirectional for client and stream objects; required for Client to server, optional for server to client
	OPTIONS = "OPTIONS"
	// Client to server for presentation and stream objects; recommended
	PAUSE = "PAUSE"
	// Client to server for presentation and stream objects; required
	PLAY = "PLAY"
	// Client to server for presentation and stream objects; optional
	RECORD = "RECORD"
	// Server to client for presentation and stream objects; optional
	REDIRECT = "REDIRECT"
	// Client to server for stream objects; required
	SETUP = "SETUP"
	// Bidirectional for presentation and stream objects; optional
	SET_PARAMETER = "SET_PARAMETER"
	// Client to server for presentation and stream objects; required
	TEARDOWN = "TEARDOWN"
)

const (
	// all requests
	Continue = 100

	// all requests
	OK = 200
	// RECORD
	Created = 201
	// RECORD
	LowOnStorageSpace = 250

	// all requests
	MultipleChoices = 300
	// all requests
	MovedPermanently = 301
	// all requests
	MovedTemporarily = 302
	// all requests
	SeeOther = 303
	// all requests
	UseProxy = 305

	// all requests
	BadRequest = 400
	// all requests
	Unauthorized = 401
	// all requests
	PaymentRequired = 402
	// all requests
	Forbidden = 403
	// all requests
	NotFound = 404
	// all requests
	MethodNotAllowed = 405
	// all requests
	NotAcceptable = 406
	// all requests
	ProxyAuthenticationRequired = 407
	// all requests
	RequestTimeout = 408
	// all requests
	Gone = 410
	// all requests
	LengthRequired = 411
	// DESCRIBE, SETUP
	PreconditionFailed = 412
	// all requests
	RequestEntityTooLarge = 413
	// all requests
	RequestURITooLong = 414
	// all requests
	UnsupportedMediaType = 415
	// SETUP
	Invalidparameter = 451
	// SETUP
	IllegalConferenceIdentifier = 452
	// SETUP
	NotEnoughBandwidth = 453
	// all requests
	SessionNotFound = 454
	// all requests
	MethodNotValidInThisState = 455
	// all requests
	HeaderFieldNotValid = 456
	// PLAY
	InvalidRange = 457
	// SET_PARAMETER
	ParameterIsReadOnly = 458
	// all requests
	AggregateOperationNotAllowed = 459
	// all requests
	OnlyAggregateOperationAllowed = 460
	// all requests
	UnsupportedTransport = 461
	// all requests
	DestinationUnreachable = 462

	// all requests
	InternalServerError = 500
	// all requests
	NotImplemented = 501
	// all requests
	BadGateway = 502
	// all requests
	ServiceUnavailable = 503
	// all requests
	GatewayTimeout = 504
	// all requests
	RTSPVersionNotSupported = 505
	// all requests
	OptionNotsupport = 551
)

func NewClient(uri string) *Client {
	return &Client{
		uri:     uri,
		timeout: time.Second * 30,
	}
}

func (c *Client) Dial() (err error) {
	c.conn = nil
	if c.URL, err = url.Parse(c.uri); err != nil {
		return err
	}
	var address string
	var hostname string // without port
	if i := strings.IndexByte(c.URL.Host, ':'); i > 0 {
		address = c.URL.Host
		hostname = c.URL.Host[:i]
	} else {
		switch c.URL.Scheme {
		case "rtsp", "rtsps", "rtspx":
			address = c.URL.Host + ":554"
		case "rtmp":
			address = c.URL.Host + ":1935"
		case "rtmps", "rtmpx":
			address = c.URL.Host + ":443"
		}
		hostname = c.URL.Host
	}
	var secure *tls.Config

	switch c.URL.Scheme {
	case "rtsp", "rtmp":
	case "rtsps", "rtspx", "rtmps", "rtmpx":
		if c.URL.Scheme[4] == 'x' || IsIP(hostname) {
			secure = &tls.Config{InsecureSkipVerify: true}
		} else {
			secure = &tls.Config{ServerName: hostname}
		}
	default:
		return errors.New("unsupported scheme: " + c.URL.Scheme)
	}
	conn, err := net.DialTimeout("tcp", address, c.timeout)
	if err != nil {
		return err
	}

	if secure == nil {
		c.conn = conn
	} else {
		tlsConn := tls.Client(conn, secure)
		if err = tlsConn.Handshake(); err != nil {
			return err
		}

		if c.URL.Scheme[4] == 'x' {
			c.URL.Scheme = c.URL.Scheme[:4] + "s"
		}
		c.conn = tlsConn
	}

	// remove UserInfo from URL
	c.auth = NewAuth(c.URL.User)
	c.URL.User = nil
	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, BufferSize)
	c.session = ""
	c.sequence = 0
	c.state = StateConn
	return nil
}

func (c *Client) Close() error {
	if c.mode == ModeActiveProducer {
		_ = c.Teardown()
	}

	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

func (c *Client) Reconnect() error {
	// close current session
	_ = c.Close()

	// start new session
	if err := c.Dial(); err != nil {
		return err
	}
	if err := c.Describe(); err != nil {
		return err
	}
	media := Media{Kind: "video", Direction: "recvonly", Codecs: []*Codec{
		{
			Name: "H264",
		},
	}}
	// restore previous medias
	if _, err := c.SetupMedia(&media); err != nil {
		return err
	}

	// for _, receiver := range c.Receivers {
	// 	if _, err := c.SetupMedia(receiver.Media); err != nil {
	// 		return err
	// 	}
	// }
	// for _, sender := range c.Senders {
	// 	if _, err := c.SetupMedia(sender.Media); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func IsIP(hostname string) bool {
	return net.ParseIP(hostname) != nil
}

func (c *Client) Options() (err error) {
	req := &Request{Method: OPTIONS, URL: c.URL}

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if val := res.Header.Get("Content-Base"); val != "" {
		c.URL, err = urlParse(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Describe() (err error) {
	// 5.3 Back channel connection
	// https://www.onvif.org/specs/stream/ONVIF-Streaming-Spec.pdf
	req := &Request{
		Method: DESCRIBE,
		URL:    c.URL,
		Header: map[string][]string{
			"Accept": {"application/sdp"},
		},
	}

	if c.Backchannel {
		req.Header.Set("Require", "www.onvif.org/ver20/backchannel")
	}

	if c.UserAgent != "" {
		// this camera will answer with 401 on DESCRIBE without User-Agent
		// https://github.com/AlexxIT/go2rtc/issues/235
		req.Header.Set("User-Agent", c.UserAgent)
	}

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if val := res.Header.Get("Content-Base"); val != "" {
		c.URL, err = urlParse(val)
		if err != nil {
			return err
		}
	}

	c.SDP = string(res.Body) // for info

	medias, err := UnmarshalSDP(res.Body)
	if err != nil {
		return err
	}

	if c.Media != "" {
		clone := make([]*Media, 0, len(medias))
		for _, media := range medias {
			if strings.Contains(c.Media, media.Kind) {
				clone = append(clone, media)
			}
		}
		medias = clone
	}

	// TODO: rewrite more smart
	if c.Medias == nil {
		c.Medias = medias
	} else if len(c.Medias) > len(medias) {
		c.Medias = c.Medias[:len(medias)]
	}

	c.mode = ModeActiveProducer

	return nil
}

func (c *Client) Announce() (err error) {
	req := &Request{
		Method: ANNOUNCE,
		URL:    c.URL,
		Header: map[string][]string{
			"Content-Type": {"application/sdp"},
		},
	}

	req.Body, err = MarshalSDP(c.SessionName, c.Medias)
	if err != nil {
		return err
	}

	_, err = c.Do(req)
	return
}

func (c *Client) Record() (err error) {
	req := &Request{
		Method: RECORD,
		URL:    c.URL,
		Header: map[string][]string{
			"Range": {"npt=0.000-"},
		},
	}

	_, err = c.Do(req)
	return
}

func (c *Client) SetupMedia(media *Media) (byte, error) {
	var transport string

	// try to use media position as channel number
	for i, m := range c.Medias {
		fmt.Printf("Media received %v [%s]\n", m.String(), m.ID)
		fmt.Printf("Media in %v [%s] %v\n", media.String(), media.ID, m.Equal(media))
		if m.Equal(media) {
			transport = fmt.Sprintf(
				// i   - RTP (data channel)
				// i+1 - RTCP (control channel)
				"RTP/AVP/TCP;unicast;interleaved=%d-%d", i*2, i*2+1,
			)
			break
		}
	}

	if transport == "" {
		return 0, fmt.Errorf("wrong media: %v", media)
	}

	rawURL := media.ID // control
	if !strings.Contains(rawURL, "://") {
		rawURL = c.URL.String()
		if !strings.HasSuffix(rawURL, "/") {
			rawURL += "/"
		}
		rawURL += media.ID
	} else if strings.HasPrefix(rawURL, "rtsp://rtsp://") {
		// fix https://github.com/AlexxIT/go2rtc/issues/830
		rawURL = rawURL[7:]
	}
	trackURL, err := urlParse(rawURL)
	if err != nil {
		return 0, err
	}

	req := &Request{
		Method: SETUP,
		URL:    trackURL,
		Header: map[string][]string{
			"Transport": {transport},
		},
	}

	res, err := c.Do(req)
	if err != nil {
		// some Dahua/Amcrest cameras fail here because two simultaneous
		// backchannel connections
		if c.Backchannel {
			c.Backchannel = false
			if err = c.Reconnect(); err != nil {
				return 0, err
			}
			return c.SetupMedia(media)
		}

		return 0, err
	}

	if c.session == "" {
		// Session: 7116520596809429228
		// Session: 216525287999;timeout=60
		if s := res.Header.Get("Session"); s != "" {
			if i := strings.IndexByte(s, ';'); i > 0 {
				c.session = s[:i]
				if i = strings.Index(s, "timeout="); i > 0 {
					c.keepalive, _ = strconv.Atoi(s[i+8:])
				}
			} else {
				c.session = s
			}
		}
	}

	// we send our `interleaved`, but camera can answer with another

	// Transport: RTP/AVP/TCP;unicast;interleaved=10-11;ssrc=10117CB7
	// Transport: RTP/AVP/TCP;unicast;destination=192.168.1.111;source=192.168.1.222;interleaved=0
	// Transport: RTP/AVP/TCP;ssrc=22345682;interleaved=0-1
	transport = res.Header.Get("Transport")
	if !strings.HasPrefix(transport, "RTP/AVP/TCP;") {
		// Escam Q6 has a bug:
		// Transport: RTP/AVP;unicast;destination=192.168.1.111;source=192.168.1.222;interleaved=0-1
		if !strings.Contains(transport, ";interleaved=") {
			return 0, fmt.Errorf("wrong transport: %s", transport)
		}
	}

	channel := Between(transport, "interleaved=", "-")
	i, err := strconv.Atoi(channel)
	if err != nil {
		return 0, err
	}

	return byte(i), nil
}

func (c *Client) Play() (err error) {
	req := &Request{Method: PLAY, URL: c.URL}
	return c.WriteRequest(req)
}

func (c *Client) Teardown() (err error) {
	// allow TEARDOWN from any state (ex. ANNOUNCE > SETUP)
	req := &Request{Method: TEARDOWN, URL: c.URL}
	return c.WriteRequest(req)
}

type Request struct {
	Method     string
	URL        *url.URL
	Proto      string
	ProtoMajor int
	ProtoMinor int
	Header     textproto.MIMEHeader
	Body       []byte
}

type Response struct {
	Status     string
	StatusCode int
	Proto      string
	Header     textproto.MIMEHeader
	Body       []byte
	Request    *Request
}

func (r *Request) String() string {
	s := r.Method + " " + r.URL.String() + " " + r.Proto + EndLine
	for k, v := range r.Header {
		s += k + ": " + v[0] + EndLine
	}
	s += EndLine
	if r.Body != nil {
		s += string(r.Body)
	}
	return s
}

func (r *Request) Write(w io.Writer) (err error) {
	_, err = w.Write([]byte(r.String()))
	return
}

func (r Response) String() string {
	s := r.Proto + " " + r.Status + EndLine
	for k, v := range r.Header {
		s += k + ": " + v[0] + EndLine
	}
	s += EndLine
	if r.Body != nil {
		s += string(r.Body)
	}
	return s
}

func (r *Response) Write(w io.Writer) (err error) {
	_, err = w.Write([]byte(r.String()))
	return
}

func (c *Client) WriteRequest(req *Request) (err error) {
	if req.Proto == "" {
		req.Proto = ProtoRTSP
	}
	if req.Header == nil {
		req.Header = make(map[string][]string)
	}
	c.sequence++
	// important to send case sensitive CSeq
	// https://github.com/AlexxIT/go2rtc/issues/7
	req.Header["CSeq"] = []string{strconv.Itoa(c.sequence)}

	c.auth.Write(req)

	if c.session != "" {
		req.Header.Set("Session", c.session)
	}

	if req.Body != nil {
		val := strconv.Itoa(len(req.Body))
		req.Header.Set("Content-Length", val)
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}

	_, err = c.conn.Write([]byte(req.String()))
	return
}

func (c *Client) ReadRequest() (*Request, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}

	tp := textproto.NewReader(c.reader)

	line, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}

	ss := strings.SplitN(line, " ", 3)
	if len(ss) != 3 {
		return nil, fmt.Errorf("wrong request: %s", line)
	}

	req := &Request{
		Method: ss[0],
		Proto:  ss[2],
	}

	req.URL, err = url.Parse(ss[1])
	if err != nil {
		return nil, err
	}

	req.Header, err = tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	if val := req.Header.Get("Content-Length"); val != "" {
		var i int
		i, err = strconv.Atoi(val)
		req.Body = make([]byte, i)
		if _, err = io.ReadAtLeast(c.reader, req.Body, i); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (c *Client) WriteResponse(res *Response) (err error) {
	if res.Proto == "" {
		res.Proto = ProtoRTSP
	}

	if res.Status == "" {
		res.Status = "200 OK"
	}

	if res.Header == nil {
		res.Header = make(map[string][]string)
	}

	if res.Request != nil && res.Request.Header != nil {
		seq := res.Request.Header.Get("CSeq")
		if seq != "" {
			res.Header.Set("CSeq", seq)
		}
	}

	if c.session != "" {
		if res.Request != nil && res.Request.Method == SETUP {
			res.Header.Set("Session", c.session+";timeout=60")
		} else {
			res.Header.Set("Session", c.session)
		}
	}

	if res.Body != nil {
		val := strconv.Itoa(len(res.Body))
		res.Header.Set("Content-Length", val)
	}

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}

	return res.Write(c.conn)
}

func (c *Client) ReadResponse() (*Response, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}
	tp := textproto.NewReader(c.reader)

	line, err := tp.ReadLine()
	if err != nil {
		return nil, err
	}
	if line == "" {
		return nil, errors.New("empty response on RTSP request")
	}

	ss := strings.SplitN(line, " ", 3)
	if len(ss) != 3 {
		return nil, fmt.Errorf("malformed response: %s", line)
	}

	res := &Response{
		Status: ss[1] + " " + ss[2],
		Proto:  ss[0],
	}

	res.StatusCode, err = strconv.Atoi(ss[1])
	if err != nil {
		return nil, err
	}

	res.Header, err = tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	if val := res.Header.Get("Content-Length"); val != "" {
		var i int
		i, err = strconv.Atoi(val)
		res.Body = make([]byte, i)
		if _, err = io.ReadAtLeast(c.reader, res.Body, i); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// Do send WriteRequest and receive and process WriteResponse
func (c *Client) Do(req *Request) (*Response, error) {
	if err := c.WriteRequest(req); err != nil {
		return nil, err
	}

	res, err := c.ReadResponse()
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		switch c.auth.Method {
		case AuthNone:
			if c.auth.ReadNone(res) {
				return c.Do(req)
			}
			return nil, errors.New("user/pass not provided")
		case AuthUnknown:
			if c.auth.Read(res) {
				return c.Do(req)
			}
		default:
			return nil, errors.New("wrong user/pass")
		}
	}

	if res.StatusCode != http.StatusOK {
		return res, fmt.Errorf("wrong response on %s", req.Method)
	}

	return res, nil
}

func (c *Client) Handle() (err error) {
	var timeout time.Duration

	var keepaliveDT time.Duration
	var keepaliveTS time.Time
	switch c.mode {
	case ModeActiveProducer:
		if c.keepalive > 5 {
			keepaliveDT = time.Duration(c.keepalive-5) * time.Second
		} else {
			keepaliveDT = 25 * time.Second
		}
		keepaliveTS = time.Now().Add(keepaliveDT)
		if c.Timeout == 0 {
			// polling frames from remote RTSP Server (ex Camera)
			timeout = time.Second * 5

			// if len(c.Receivers) == 0
			if false {
				// if we only send audio to camera
				// https://github.com/AlexxIT/go2rtc/issues/659
				timeout += keepaliveDT
			}
		} else {
			timeout = time.Second * time.Duration(c.Timeout)
		}
	case ModePassiveProducer:
		// polling frames from remote RTSP Client (ex FFmpeg)
		if c.Timeout == 0 {
			timeout = time.Second * 15
		} else {
			timeout = time.Second * time.Duration(c.Timeout)
		}
	case ModePassiveConsumer:
		// pushing frames to remote RTSP Client (ex VLC)
		timeout = time.Second * 60
	default:
		return fmt.Errorf("wrong RTSP conn mode: %d", c.mode)
	}

	for c.state != StateNone {
		ts := time.Now()

		if err = c.conn.SetReadDeadline(ts.Add(timeout)); err != nil {
			return
		}

		// we can read:
		// 1. RTP interleaved: `$` + 1B channel number + 2B size
		// 2. RTSP response:   RTSP/1.0 200 OK
		// 3. RTSP request:    OPTIONS ...
		var buf4 []byte // `$` + 1B channel number + 2B size
		buf4, err = c.reader.Peek(4)
		if err != nil {
			fmt.Printf("error %v", err)
			return
		}

		var channelID byte
		var size uint16

		if buf4[0] != '$' {
			switch string(buf4) {
			case "RTSP":
				// var res *Response
				if _, err = c.ReadResponse(); err != nil {
					return
				}

				// for playing backchannel only after OK response on play
				c.playOK = true
				continue

			case "OPTI", "TEAR", "DESC", "SETU", "PLAY", "PAUS", "RECO", "ANNO", "GET_", "SET_":
				var req *Request
				if req, err = c.ReadRequest(); err != nil {
					return
				}

				if req.Method == OPTIONS {
					res := &Response{Request: req}
					if err = c.WriteResponse(res); err != nil {
						return
					}
				}
				continue

			default:

				for i := 0; ; i++ {
					// search next start symbol
					if _, err = c.reader.ReadBytes('$'); err != nil {
						return err
					}

					if channelID, err = c.reader.ReadByte(); err != nil {
						return err
					}

					// TODO: better check maximum good channel ID
					if channelID >= 20 {
						continue
					}

					buf4 = make([]byte, 2)
					if _, err = io.ReadFull(c.reader, buf4); err != nil {
						return err
					}

					// check if size good for RTP
					size = binary.BigEndian.Uint16(buf4)
					if size <= 1500 {
						break
					}

					// 10 tries to find good packet
					if i >= 10 {
						return fmt.Errorf("RTSP wrong input")
					}
				}
			}
		} else {
			// hope that the odd channels are always RTCP
			channelID = buf4[1]

			// get data size
			size = binary.BigEndian.Uint16(buf4[2:])

			// skip 4 bytes from c.reader.Peek
			if _, err = c.reader.Discard(4); err != nil {
				return
			}
		}

		// init memory for data
		buf := make([]byte, size)
		if _, err = io.ReadFull(c.reader, buf); err != nil {
			return
		}

		if channelID&1 == 0 {
			packet := &rtp.Packet{}
			if err = packet.Unmarshal(buf); err != nil {
				return
			}
			fmt.Println(packet)

			// for _, receiver := range c.Receivers {
			// 	if receiver.ID == channelID {
			// 		receiver.WriteRTP(packet)
			// 		break
			// 	}
			// }
		} else {
			msg := &RTCP{Channel: channelID}

			if err = msg.Header.Unmarshal(buf); err != nil {
				continue
			}

			msg.Packets, err = rtcp.Unmarshal(buf)
			if err != nil {
				continue
			}

		}

		if keepaliveDT != 0 && ts.After(keepaliveTS) {
			req := &Request{Method: OPTIONS, URL: c.URL}
			if err = c.WriteRequest(req); err != nil {
				return
			}

			keepaliveTS = ts.Add(keepaliveDT)
		}
	}

	return
}
