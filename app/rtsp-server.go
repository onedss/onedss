package app

import (
	"fmt"
	"github.com/onedss/onedss/rtsp"
	"github.com/onedss/onedss/utils"
	"log"
)

type rtsp_server struct {
	rtspPort   int
	rtspServer *rtsp.Server
}

func NewOneRtspServer(rtspPort int, rtspServer *rtsp.Server) (server *rtsp_server) {
	return &rtsp_server{
		rtspPort:   rtspPort,
		rtspServer: rtspServer,
	}
}

func (p *rtsp_server) Start() (err error) {
	if p.rtspServer == nil {
		err = fmt.Errorf("RTSP Server Not Found")
		return
	}
	sport := ""
	if p.rtspPort != 554 {
		sport = fmt.Sprintf(":%d", p.rtspPort)
	}
	link := fmt.Sprintf("rtsp://%s%s", utils.LocalIP(), sport)
	log.Println("rtsp server start -->", link)
	go func() {
		if err := p.rtspServer.Start(); err != nil {
			log.Println("start rtsp server error", err)
		}
		log.Println("rtsp server end")
	}()
	return
}

func (p *rtsp_server) Stop() (err error) {
	if p.rtspServer == nil {
		err = fmt.Errorf("RTSP Server Not Found")
		return
	}
	p.rtspServer.Stop()
	return
}

func (p *rtsp_server) GetPort() int {
	return p.rtspPort
}
