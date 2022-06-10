package app

import (
	"context"
	"fmt"
	"github.com/onedss/onedss/routers"
	"github.com/onedss/onedss/utils"
	"log"
	"net/http"
	"time"
)

type http_server struct {
	httpPort   int
	httpServer *http.Server
}

func NewOneHttpServer(httpPort int) (server *http_server) {
	return &http_server{
		httpPort: httpPort,
	}
}

func (p *http_server) Start() (err error) {
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

func (p *http_server) Stop() (err error) {
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

func (p *http_server) GetPort() int {
	return p.httpPort
}
