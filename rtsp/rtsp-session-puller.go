package rtsp

import (
	"github.com/onedss/EasyGoLib/utils"
	"github.com/teris-io/shortid"
	"log"
	"time"
)

type SessionPuller struct {
	*Session
	RTSPClient *RTSPClient
}

func NewSessionPuller(server *Server, client *RTSPClient) *SessionPuller {
	session := &Session{
		ID:      shortid.MustGenerate(),
		Server:  server,
		StartAt: time.Now(),
		Timeout: utils.Conf().Section("rtsp").Key("timeout").MustInt(0),

		RTPHandles:  make([]func(*RTPPack), 0),
		StopHandles: make([]func(), 0),
	}
	puller := &SessionPuller{
		Session:    session,
		RTSPClient: client,
	}
	return puller
}

func (puller *SessionPuller) AddRTPHandles(f func(*RTPPack)) {
	puller.RTPHandles = append(puller.RTPHandles, f)
}

func (puller *SessionPuller) AddStopHandles(f func()) {
	puller.StopHandles = append(puller.StopHandles, f)
}

func (puller *SessionPuller) GetID() string {
	return puller.Session.ID
}

func (puller *SessionPuller) GetPath() string {
	return puller.Session.Path
}

func (puller *SessionPuller) Stop() {
	log.Println("Puller Stopped :", puller.ID)
	if puller.Stoped {
		return
	}
	puller.Stoped = true
	for _, h := range puller.StopHandles {
		h()
	}
	if puller.privateConn != nil {
		puller.connRW.Flush()
		puller.privateConn.Close()
		puller.privateConn = nil
	}
	if puller.UDPClient != nil {
		puller.UDPClient.Stop()
		puller.UDPClient = nil
	}
	if puller.RTSPClient != nil {
		puller.RTSPClient.Stop()
		puller.RTSPClient = nil
	}
}

func (puller *SessionPuller) Start() {
	client := puller.RTSPClient
	if !client.InitFlag {
		log.Printf("Pull to push fail.")
		return
	}
	if client.CustomPath != "" {
		puller.Path = client.CustomPath
	} else {
		puller.Path = client.Path
	}
	puller.URL = client.URL
	puller.SDPRaw = client.SDPRaw
	puller.SDPMap = ParseSDP(client.SDPRaw)
	sdp, ok := puller.SDPMap["audio"]
	if ok {
		puller.AControl = sdp.Control
		puller.ACodec = sdp.Codec
		log.Printf("audio codec[%s]\n", puller.ACodec)
	}
	sdp, ok = puller.SDPMap["video"]
	if ok {
		puller.VControl = sdp.Control
		puller.VCodec = sdp.Codec
		log.Printf("video codec[%s]\n", puller.VCodec)
	}
	pusher := NewPusher(puller.Session)
	pusher.StopHandles = append(pusher.StopHandles, func() {
		puller.Stop()
	})
	client.RTPHandles = append(client.RTPHandles, func(pack *RTPPack) {
		pusher.QueueRTP(pack)
		pusher.InBytes += pack.Buffer.Len()
	})
	client.StopHandles = append(client.StopHandles, func() {
		pusher.Stoped = true
		pusher.ClearPlayer()
		pusher.GetServer().RemovePusher(pusher)
		pusher.cond.Broadcast()
	})
	client.Start()
	puller.Server.AddPusher(pusher)
}
