package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	figure "github.com/common-nighthawk/go-figure"
	"github.com/onedss/onedss/models"
	"github.com/onedss/onedss/routers"

	"github.com/onedss/EasyGoLib/utils"
	"github.com/onedss/onedss/rtsp"
	"github.com/onedss/onedss/service"
)

type program struct {
	httpPort   int
	httpServer *http.Server
	rtspPort   int
	rtspServer *rtsp.Server
}

func (p *program) StopHTTP() (err error) {
	if p.httpServer == nil {
		err = fmt.Errorf("HTTP Server Not Found")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = p.httpServer.Shutdown(ctx); err != nil {
		return
	}
	return
}

func (p *program) StartHTTP() (err error) {
	p.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", p.httpPort),
		Handler:           routers.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	link := fmt.Sprintf("http://%s:%d", utils.LocalIP(), p.httpPort)
	log.Println("http server start -->", link)
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("start http server error", err)
		}
		log.Println("http server end")
	}()
	return
}

func (p *program) StartRTSP() (err error) {
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

func (p *program) StopRTSP() (err error) {
	if p.rtspServer == nil {
		err = fmt.Errorf("RTSP Server Not Found")
		return
	}
	p.rtspServer.Stop()
	return
}

func (p *program) Start(s service.Service) (err error) {
	log.Println("********** START **********")
	if utils.IsPortInUse(p.httpPort) {
		err = fmt.Errorf("HTTP port[%d] In Use", p.httpPort)
		return
	}
	if utils.IsPortInUse(p.rtspPort) {
		err = fmt.Errorf("RTSP port[%d] In Use", p.rtspPort)
		return
	}
	err = models.Init()
	if err != nil {
		return
	}
	err = routers.Init()
	if err != nil {
		return
	}
	p.StartRTSP()
	p.StartHTTP()
	if !utils.Debug {
		log.Println("log files -->", utils.LogDir())
		log.SetOutput(utils.GetLogWriter())
	}
	go func() {
		for range routers.API.RestartChan {
			p.StopHTTP()
			p.StopRTSP()
			utils.ReloadConf()
			p.StartRTSP()
			p.StartHTTP()
		}
	}()
	return
}

func (p *program) Stop(s service.Service) (err error) {
	defer log.Println("********** STOP **********")
	defer utils.CloseLogWriter()
	p.StopHTTP()
	p.StopRTSP()
	models.Close()
	return
}

func main() {
	log.SetPrefix("[EasyDarwin] ")
	log.SetFlags(log.LstdFlags)
	if utils.Debug {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	}
	sec := utils.Conf().Section("service")
	svcConfig := &service.Config{
		Name:        sec.Key("name").MustString("EasyDarwin_Service"),
		DisplayName: sec.Key("display_name").MustString("EasyDarwin_Service"),
		Description: sec.Key("description").MustString("EasyDarwin_Service"),
	}

	httpPort := utils.Conf().Section("http").Key("port").MustInt(10008)
	rtspServer := rtsp.GetServer()
	p := &program{
		httpPort:   httpPort,
		rtspPort:   rtspServer.TCPPort,
		rtspServer: rtspServer,
	}
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
