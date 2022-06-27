package app

import (
	"github.com/common-nighthawk/go-figure"
	"github.com/onedss/onedss/rtsp"
	"github.com/onedss/onedss/service"
	"github.com/onedss/onedss/utils"
	"log"
	"os"
)

func StartApp() {
	log.Println("config file -->", utils.ConfFile())
	sec := utils.Conf().Section("service")
	svcConfig := &service.Config{
		Name:        sec.Key("name").MustString("EasyDarwin_Service"),
		DisplayName: sec.Key("display_name").MustString("EasyDarwin_Service"),
		Description: sec.Key("description").MustString("EasyDarwin_Service"),
	}

	httpPort := utils.Conf().Section("http").Key("port").MustInt(10008)
	oneHttpServer := NewOneHttpServer(httpPort)
	sigPort := utils.Conf().Section("signaling").Key("port").MustInt(1989)
	signalingServer := NewSignalingServer(sigPort)
	rtspServer := rtsp.GetServer()
	oneRtspServer := NewOneRtspServer(rtspServer.TCPPort, rtspServer)
	p := &application{}
	p.AddServer(oneHttpServer)
	p.AddServer(signalingServer)
	p.AddServer(oneRtspServer)
	var s, err = service.New(p, svcConfig)
	if err != nil {
		log.Println(err)
		utils.PauseExit()
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "install" || os.Args[1] == "stop" {
			figure.NewFigure("EasyDarwin", "", false).Print()
		}
		log.Println(svcConfig.Name, os.Args[1], "...")
		if err = service.Control(s, os.Args[1]); err != nil {
			log.Println(err)
			utils.PauseExit()
		}
		log.Println(svcConfig.Name, os.Args[1], "ok")
		return
	}
	figure.NewFigure("EasyDarwin", "", false).Print()
	if err = s.Run(); err != nil {
		log.Println(err)
		utils.PauseExit()
	}
}
