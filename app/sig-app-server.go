package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/onedss/onedss/signaling"
	"github.com/onedss/onedss/utils"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type signaling_server struct {
	httpPort   int
	httpServer *http.Server
	rooms      sync.Map
	mux        *http.ServeMux
}

func NewSignalingServer(httpPort int) (server *signaling_server) {
	return &signaling_server{
		httpPort: httpPort,
	}
}

func (p *signaling_server) Start() (err error) {
	p.mux = http.NewServeMux()
	p.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", p.httpPort),
		Handler:           p.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	wwwDir := filepath.Join(utils.DataDir(), "www-sig")
	log.Println("www-sig root -->", wwwDir)
	p.mux.Handle("/", http.FileServer(http.Dir(wwwDir)))

	link := fmt.Sprintf("http://%s:%d", utils.LocalIP(), p.httpPort)
	log.Println("signaling server start -->", link)

	// Key is name of room, value is Room
	p.mux.Handle("/sig/v1/rtc", websocket.Handler(func(c *websocket.Conn) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		r := c.Request()
		log.Printf("Serve client %v at %v", r.RemoteAddr, r.RequestURI)
		defer c.Close()

		var self *signaling.Participant
		go func() {
			<-ctx.Done()
			if self == nil {
				return
			}

			// Notify other peers that we're quiting.
			// @remark The ctx(of self) is done, so we must use a new context.
			go self.Room.Notify(context.Background(), self, "leave", "", "")

			self.Room.Remove(self)
			log.Printf("Remove client %v", self)
		}()

		inMessages := make(chan []byte, 0)
		go func() {
			defer cancel()

			buf := make([]byte, 16384)
			for {
				n, err := c.Read(buf)
				if err != nil {
					log.Printf("Ignore err %v for %v", err, r.RemoteAddr)
					break
				}

				select {
				case <-ctx.Done():
				case inMessages <- buf[:n]:
				}
			}
		}()

		outMessages := make(chan []byte, 0)
		go func() {
			defer cancel()

			handleMessage := func(m []byte) error {
				action := struct {
					TID     string `json:"tid"`
					Message struct {
						Action string `json:"action"`
					} `json:"msg"`
				}{}
				if err := json.Unmarshal(m, &action); err != nil {
					return errors.Wrapf(err, "Unmarshal %s", m)
				}

				var res interface{}
				if action.Message.Action == "join" {
					obj := struct {
						Message struct {
							Room    string `json:"room"`
							Display string `json:"display"`
						} `json:"msg"`
					}{}
					if err := json.Unmarshal(m, &obj); err != nil {
						return errors.Wrapf(err, "Unmarshal %s", m)
					}

					r, _ := p.rooms.LoadOrStore(obj.Message.Room, &signaling.Room{Name: obj.Message.Room})
					p := &signaling.Participant{Room: r.(*signaling.Room), Display: obj.Message.Display, Out: outMessages}
					if err := r.(*signaling.Room).Add(p); err != nil {
						return errors.Wrapf(err, "join")
					}

					self = p
					log.Printf("Join %v ok", self)

					res = struct {
						Action       string                   `json:"action"`
						Room         string                   `json:"room"`
						Self         *signaling.Participant   `json:"self"`
						Participants []*signaling.Participant `json:"participants"`
					}{
						action.Message.Action, obj.Message.Room, p, r.(*signaling.Room).Participants,
					}

					go r.(*signaling.Room).Notify(ctx, p, action.Message.Action, "", "")
				} else if action.Message.Action == "publish" {
					obj := struct {
						Message struct {
							Room    string `json:"room"`
							Display string `json:"display"`
						} `json:"msg"`
					}{}
					if err := json.Unmarshal(m, &obj); err != nil {
						return errors.Wrapf(err, "Unmarshal %s", m)
					}

					r, _ := p.rooms.LoadOrStore(obj.Message.Room, &signaling.Room{Name: obj.Message.Room})
					p := r.(*signaling.Room).Get(obj.Message.Display)

					// Now, the peer is publishing.
					p.Publishing = true

					go r.(*signaling.Room).Notify(ctx, p, action.Message.Action, "", "")
				} else if action.Message.Action == "control" {
					obj := struct {
						Message struct {
							Room    string `json:"room"`
							Display string `json:"display"`
							Call    string `json:"call"`
							Data    string `json:"data"`
						} `json:"msg"`
					}{}
					if err := json.Unmarshal(m, &obj); err != nil {
						return errors.Wrapf(err, "Unmarshal %s", m)
					}

					r, _ := p.rooms.LoadOrStore(obj.Message.Room, &signaling.Room{Name: obj.Message.Room})
					p := r.(*signaling.Room).Get(obj.Message.Display)

					go r.(*signaling.Room).Notify(ctx, p, action.Message.Action, obj.Message.Call, obj.Message.Data)
				} else {
					return errors.Errorf("Invalid message %s", m)
				}

				if b, err := json.Marshal(struct {
					TID     string      `json:"tid"`
					Message interface{} `json:"msg"`
				}{
					action.TID, res,
				}); err != nil {
					return errors.Wrapf(err, "marshal")
				} else {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case outMessages <- b:
					}
				}

				return nil
			}

			for m := range inMessages {
				if err := handleMessage(m); err != nil {
					log.Printf("Handle %s err %v", m, err)
					break
				}
			}
		}()

		for m := range outMessages {
			if _, err := c.Write(m); err != nil {
				log.Printf("Ignore err %v for %v", err, r.RemoteAddr)
				break
			}
		}
	}))
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("start signaling server error", err)
		}
		log.Println("signaling server end")
	}()
	return nil
}

func (p *signaling_server) Stop() (err error) {
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

func (p *signaling_server) GetPort() int {
	return p.httpPort
}
