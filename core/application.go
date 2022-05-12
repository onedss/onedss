package core

import "github.com/onedss/onedss/service"

type application struct {
	servers []OneServer
}

func (p *application) Start(s service.Service) (err error) {
	for _, server := range p.servers {
		err := server.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *application) Stop(s service.Service) (err error) {
	for _, server := range p.servers {
		server.Stop()
	}
	return nil
}

func (p *application) AddServer(server OneServer) {
	p.servers = append(p.servers, server)
}
