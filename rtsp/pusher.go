package rtsp

import (
	"fmt"
	"github.com/onedss/onedss/core"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/onedss/onedss/utils"
)

type Pusher struct {
	core.SessionLogger
	*Session

	players        map[string]*Player //SessionID <-> Player
	playersLock    sync.RWMutex
	gopCacheEnable bool
	gopCache       []*RTPPack
	gopCacheLock   sync.RWMutex
	UDPServer      *UDPServer

	cond  *sync.Cond
	queue []*RTPPack
}

func NewPusher(session *Session) (pusher *Pusher) {
	pusher = &Pusher{
		Session:        session,
		players:        make(map[string]*Player),
		gopCacheEnable: utils.Conf().Section("rtsp").Key("gop_cache_enable").MustBool(true),
		gopCache:       make([]*RTPPack, 0),

		cond:  sync.NewCond(&sync.Mutex{}),
		queue: make([]*RTPPack, 0),
	}
	pusher.SetLogger(log.New(os.Stdout, fmt.Sprintf("[%s] ", session.ID), log.LstdFlags|log.Lshortfile|log.Lmicroseconds))
	if !utils.Debug {
		pusher.GetLogger().SetOutput(utils.GetLogWriter())
	}
	session.AddRTPHandles(func(pack *RTPPack) {
		pusher.QueueRTP(pack)
	})
	session.AddStopHandles(func() {
		pusher.Server.RemovePusher(pusher)
		pusher.cond.Broadcast()
		if pusher.UDPServer != nil {
			pusher.UDPServer.Stop()
			pusher.UDPServer = nil
		}
	})
	return
}

func (pusher *Pusher) GetServer() *Server {
	return pusher.Server
}

func (pusher *Pusher) GetPath() string {
	return pusher.Path
}

func (pusher *Pusher) GetID() string {
	return pusher.ID
}

func (pusher *Pusher) GetVCodec() string {
	return pusher.VCodec
}

func (pusher *Pusher) GetACodec() string {
	return pusher.ACodec
}

func (pusher *Pusher) GetAControl() string {
	return pusher.AControl
}

func (pusher *Pusher) GetVControl() string {
	return pusher.VControl
}

func (pusher *Pusher) GetSDPRaw() string {
	return pusher.SDPRaw
}

func (pusher *Pusher) GetUDPServer() *UDPServer {
	return pusher.UDPServer
}

func (pusher *Pusher) SetUDPServer(udpServer *UDPServer) {
	pusher.UDPServer = udpServer
}

func (pusher *Pusher) GetInBytes() int {
	return pusher.InBytes
}

func (pusher *Pusher) GetOutBytes() int {
	return pusher.OutBytes
}

func (pusher *Pusher) GetTransType() string {
	return pusher.TransType.String()
}

func (pusher *Pusher) GetStartAt() time.Time {
	return pusher.StartAt
}

func (pusher *Pusher) GetSource() string {
	return pusher.URL
}

func (pusher *Pusher) QueueRTP(pack *RTPPack) *Pusher {
	pusher.cond.L.Lock()
	pusher.queue = append(pusher.queue, pack)
	pusher.cond.Signal()
	pusher.cond.L.Unlock()
	return pusher
}

func (pusher *Pusher) Start() {
	logger := pusher.GetLogger()
	logger.Printf("Pusher Start Begin. [%s]", pusher.ID)
	for !pusher.Stoped {
		var pack *RTPPack
		pusher.cond.L.Lock()
		if len(pusher.queue) == 0 {
			pusher.cond.Wait()
		}
		if len(pusher.queue) > 0 {
			pack = pusher.queue[0]
			pusher.queue = pusher.queue[1:]
		}
		pusher.cond.L.Unlock()
		if pack == nil {
			if !pusher.Stoped {
				logger.Printf("pusher not stoped, but queue take out nil pack")
			}
			continue
		}
		if pusher.UDPSender != nil && pack.Type == RTP_TYPE_AUDIO {
			pusher.UDPSender.Write(pack.Buffer.Bytes())
			//src := pack.Buffer.Bytes()
			//encodedStr := hex.EncodeToString(src)
			//logger.Println(encodedStr)
		}
		if pusher.gopCacheEnable && pack.Type == RTP_TYPE_VIDEO {
			pusher.gopCacheLock.Lock()
			if strings.EqualFold(pusher.VCodec, "h264") {
				if rtp := ParseRTP(pack.Buffer.Bytes()); rtp != nil && rtp.IsKeyframeStart() {
					pusher.gopCache = make([]*RTPPack, 0)
				}
				pusher.gopCache = append(pusher.gopCache, pack)
			} else if strings.EqualFold(pusher.VCodec, "h265") {
				if rtp := ParseRTP(pack.Buffer.Bytes()); rtp != nil && rtp.IsKeyframeStartH265() {
					pusher.gopCache = make([]*RTPPack, 0)
				}
				pusher.gopCache = append(pusher.gopCache, pack)
			}
			pusher.gopCacheLock.Unlock()
		}

		pusher.BroadcastRTP(pack)
	}
	logger.Printf("Pusher Star End. [%s]", pusher.ID)
}

//func (pusher *Pusher) Stop() {
//	if pusher.Session != nil {
//		pusher.Session.Stop()
//		return
//	}
//	pusher.Stoped = true
//}

func (pusher *Pusher) BroadcastRTP(pack *RTPPack) *Pusher {
	for _, player := range pusher.GetPlayers() {
		player.QueueRTP(pack)
		pusher.OutBytes += pack.Buffer.Len()
	}
	return pusher
}

func (pusher *Pusher) GetPlayers() (players map[string]*Player) {
	players = make(map[string]*Player)
	pusher.playersLock.RLock()
	for k, v := range pusher.players {
		players[k] = v
	}
	pusher.playersLock.RUnlock()
	return
}

func (pusher *Pusher) AddPlayer(player *Player) *Pusher {
	logger := pusher.GetLogger()
	if pusher.gopCacheEnable {
		pusher.gopCacheLock.RLock()
		for _, pack := range pusher.gopCache {
			player.QueueRTP(pack)
			pusher.OutBytes += pack.Buffer.Len()
		}
		pusher.gopCacheLock.RUnlock()
	}

	pusher.playersLock.Lock()
	if _, ok := pusher.players[player.ID]; !ok {
		pusher.players[player.ID] = player
		go player.Start()
		logger.Printf("%v start, now player size[%d]", player, len(pusher.players))
	}
	pusher.playersLock.Unlock()
	return pusher
}

func (pusher *Pusher) RemovePlayer(player *Player) *Pusher {
	logger := pusher.GetLogger()
	pusher.playersLock.Lock()
	delete(pusher.players, player.ID)
	logger.Printf("%v end, now player size[%d]\n", player, len(pusher.players))
	pusher.playersLock.Unlock()
	return pusher
}

func (pusher *Pusher) ClearPlayer() {
	// copy a new map to avoid deadlock
	pusher.playersLock.Lock()
	players := pusher.players
	pusher.players = make(map[string]*Player)
	pusher.playersLock.Unlock()
	go func() { // do not block
		for _, v := range players {
			v.Stop()
		}
	}()
}
