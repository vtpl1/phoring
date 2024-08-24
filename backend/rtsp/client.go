package rtsp

func NewClient(uri string) *Conn {
	return &Conn{
		uri: uri,
	}
}

func (c *Conn) Dial() (err error) {
	return nil
}
