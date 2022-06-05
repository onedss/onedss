package rtsp

import (
	"encoding/hex"
	"github.com/onedss/EasyGoLib/utils"
	"github.com/teris-io/shortid"
	"log"
	"sync"
	"time"
)

type SessionPuller struct {
	*Session
	RTSPClient *RTSPClient
	path       string
}

func NewSessionPuller(server *Server, client *RTSPClient) *SessionPuller {
	//networkBuffer := utils.Conf().Section("rtsp").Key("network_buffer").MustInt(1048576)
	session := &Session{
		ID:     shortid.MustGenerate(),
		Server: server,
		//Conn:    conn,
		//connRW:  bufio.NewReadWriter(bufio.NewReaderSize(conn, networkBuffer), bufio.NewWriterSize(conn, networkBuffer)),
		StartAt: time.Now(),
		Timeout: utils.Conf().Section("rtsp").Key("timeout").MustInt(0),

		RTPHandles:  make([]func(*RTPPack), 0),
		StopHandles: make([]func(), 0),
	}
	puller := &SessionPuller{
		Session:    session,
		RTSPClient: client,
		path:       client.CustomPath,
	}
	return puller
}

func (puller *SessionPuller) ID() string {
	return puller.Session.ID
}

func (puller *SessionPuller) Path() string {
	return puller.path
}

func (puller *SessionPuller) Stop() {
	log.Println("Stop :", puller.ID)
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
	pusher := &Pusher{
		//RTSPServer:     puller.Server,
		//RTSPClient:     puller.RTSPClient,
		Session:        puller.Session,
		players:        make(map[string]*Player),
		gopCacheEnable: utils.Conf().Section("rtsp").Key("gop_cache_enable").MustBool(true),
		gopCache:       make([]*RTPPack, 0),

		cond:  sync.NewCond(&sync.Mutex{}),
		queue: make([]*RTPPack, 0),
	}
	client := puller.RTSPClient
	client.RTPHandles = append(client.RTPHandles, func(pack *RTPPack) {
		pusher.QueueRTP(pack)
		encodedStr := hex.EncodeToString(pack.Buffer.Bytes())
		log.Println(encodedStr)
	})
	client.StopHandles = append(client.StopHandles, func() {
		pusher.ClearPlayer()
		pusher.GetServer().RemovePusher(pusher)
		pusher.cond.Broadcast()
	})
	client.Start()
	puller.Server.AddPusher(pusher)
}
