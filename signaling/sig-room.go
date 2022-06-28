package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"sync"
)

type Participant struct {
	Room       *Room       `json:"-"`
	Display    string      `json:"display"`
	Publishing bool        `json:"publishing"`
	Out        chan []byte `json:"-"`
}

func (v *Participant) String() string {
	return fmt.Sprintf("display=%v, room=%v", v.Display, v.Room.Name)
}

type Room struct {
	Name         string         `json:"room"`
	Participants []*Participant `json:"participants"`
	lock         sync.RWMutex   `json:"-"`
}

func (v *Room) String() string {
	return fmt.Sprintf("room=%v, participants=%v", v.Name, len(v.Participants))
}

func (v *Room) Add(p *Participant) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	for _, r := range v.Participants {
		if r.Display == p.Display {
			return errors.Errorf("Participant %v exists in room %v", p.Display, v.Name)
		}
	}

	v.Participants = append(v.Participants, p)
	return nil
}

func (v *Room) Get(display string) *Participant {
	v.lock.RLock()
	defer v.lock.RUnlock()

	for _, r := range v.Participants {
		if r.Display == display {
			return r
		}
	}

	return nil
}

func (v *Room) Remove(p *Participant) {
	v.lock.Lock()
	defer v.lock.Unlock()

	for i, r := range v.Participants {
		if p == r {
			v.Participants = append(v.Participants[:i], v.Participants[i+1:]...)
			return
		}
	}
}

func (v *Room) Notify(ctx context.Context, peer *Participant, event, param, data string) {
	var participants []*Participant
	func() {
		v.lock.RLock()
		defer v.lock.RUnlock()
		participants = append(participants, v.Participants...)
	}()

	for _, r := range participants {
		if r == peer {
			continue
		}

		res := struct {
			Action       string         `json:"action"`
			Event        string         `json:"event"`
			Param        string         `json:"param,omitempty"`
			Data         string         `json:"data,omitempty"`
			Room         string         `json:"room"`
			Self         *Participant   `json:"self"`
			Peer         *Participant   `json:"peer"`
			Participants []*Participant `json:"participants"`
		}{
			"notify", event, param, data,
			v.Name, r, peer, participants,
		}

		b, err := json.Marshal(struct {
			Message interface{} `json:"msg"`
		}{
			res,
		})
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case r.Out <- b:
		}

		log.Printf("Notify %v about %v %v", r, peer, event)
	}
}
