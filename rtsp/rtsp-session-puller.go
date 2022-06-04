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
	log.Println(puller.ID)
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
}

func (puller *SessionPuller) Start() {
	defer puller.Stop()

}
