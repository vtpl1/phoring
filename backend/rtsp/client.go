package rtsp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	URL *url.URL

	sequence int
	auth     *Auth
	conn     net.Conn
	session  string
	uri      string
	timeout  time.Duration
}

var (
	errNotImplemented = errors.New("not implemented")
)

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
		return err
	}
	tlsConn := tls.Client(conn, secure)
	if err = tlsConn.Handshake(); err != nil {
		return err
	}

	if c.URL.Scheme[4] == 'x' {
		c.URL.Scheme = c.URL.Scheme[:4] + "s"
	}
	c.conn = tlsConn

	// remove UserInfo from URL
	c.auth = NewAuth(c.URL.User)
	c.URL.User = nil
	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, c.BufferSize)
	c.session = ""
	c.sequence = 0
	c.state = StateConn
	return nil
}

func IsIP(hostname string) bool {
	return net.ParseIP(hostname) != nil
}

func (c *Client) Options() (err error) {
	return errNotImplemented
}

func (c *Client) Describe() (err error) {
	return errNotImplemented
}

func (c *Client) Announce() (err error) {
	return errNotImplemented
}

func (c *Client) Record() (err error) {
	return errNotImplemented
}

func (c *Client) SetupMedia() (err error) {
	return errNotImplemented
}

func (c *Client) Play() (err error) {
	return errNotImplemented
}

func (c *Client) Teardown() (err error) {
	return errNotImplemented
}

type Request struct {
	Method        string
	URL           *url.URL
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        http.Header
	ContentLength int
	Body          io.ReadCloser
}

type Response struct {
	Proto string

	StatusCode int
	Status     string

	ContentLength int64

	Header http.Header
	Body   io.ReadCloser
	// Body   []byte
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
		str, _ := io.ReadAll(req.Body)
		val := strconv.Itoa(len(str))
		req.Header.Set("Content-Length", val)
	}
	if err := c.conn.SetWriteDeadline(time.Now().Add(Timeout)); err != nil {
		return err
	}
	_, err = c.conn.Write([]byte(req.String()))
	return
}

func (c *Client) ReadRequest() (*Request, error) {
	return nil, errNotImplemented
}

func (c *Client) WriteResponse(res *Response) (err error) {
	return errNotImplemented
}

func (c *Client) ReadResponse() (*Response, error) {
	return nil, errNotImplemented
}

// Do send WriteRequest and receive and process WriteResponse
func (c *Client) Do(req *Request) (*Response, error) {
	return nil, errNotImplemented
}

func (c *Client) Handle() (err error) {
	return errNotImplemented
}
