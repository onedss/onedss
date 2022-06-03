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
	//networkBuffer := utils.Conf().Section("rtsp").Key("network_buffer").MustInt(1048576)
	s := &Session{
		ID:     shortid.MustGenerate(),
		Server: server,
		//Conn:    conn,
		//connRW:  bufio.NewReadWriter(bufio.NewReaderSize(conn, networkBuffer), bufio.NewWriterSize(conn, networkBuffer)),
		StartAt: time.Now(),
		Timeout: utils.Conf().Section("rtsp").Key("timeout").MustInt(0),

		RTPHandles:  make([]func(*RTPPack), 0),
		StopHandles: make([]func(), 0),
	}
	session := &SessionPuller{
		Session:    s,
		RTSPClient: client,
	}
	return session
}

func (session *SessionPuller) Stop() {
	log.Println(session.ID)
	if session.Stoped {
		return
	}
	session.Stoped = true
	for _, h := range session.StopHandles {
		h()
	}
	if session.privateConn != nil {
		session.connRW.Flush()
		session.privateConn.Close()
		session.privateConn = nil
	}
	if session.UDPClient != nil {
		session.UDPClient.Stop()
		session.UDPClient = nil
	}
}

func (session *SessionPuller) Start() {
	defer session.Stop()

}
