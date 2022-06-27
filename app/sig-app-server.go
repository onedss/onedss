package app

import "net/http"

type signaling_server struct {
	httpPort   int
	httpServer *http.Server
}

func NewSignalingServer(httpPort int) (server *signaling_server) {
	return &signaling_server{
		httpPort: httpPort,
	}
}

func (p *signaling_server) Start() (err error) {
	return nil
}

func (p *signaling_server) Stop() (err error) {
	return nil
}

func (p *signaling_server) GetPort() int {
	return p.httpPort
}
