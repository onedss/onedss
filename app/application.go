package app

import (
	"fmt"
	"github.com/onedss/onedss/core"
	"github.com/onedss/onedss/models"
	"github.com/onedss/onedss/routers"
	"github.com/onedss/onedss/service"
	"github.com/onedss/onedss/utils"
	"log"
)

type application struct {
	servers []core.OneServer
}

func (p *application) Start(s service.Service) (err error) {
	log.Println("********** START **********")
	for _, server := range p.servers {
		port := server.GetPort()
		if utils.IsPortInUse(port) {
			err = fmt.Errorf("TCP port[%d] In Use", port)
			return
		}
	}
	err = models.Init()
	if err != nil {
		return
	}
	err = routers.Init()
	if err != nil {
		return
	}
	for _, server := range p.servers {
		err := server.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *application) Stop(s service.Service) (err error) {
	defer log.Println("********** STOP **********")
	defer utils.CloseLogWriter()
	for _, server := range p.servers {
		server.Stop()
	}
	models.Close()
	return
}

func (p *application) AddServer(server core.OneServer) {
	p.servers = append(p.servers, server)
}
