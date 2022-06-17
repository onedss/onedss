package rtsp

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onedss/onedss/utils"

	"github.com/teris-io/shortid"
)

type SessionType int

const (
	SESSION_TYPE_PUSHER SessionType = iota
	SESSEION_TYPE_PLAYER
)

func (st SessionType) String() string {
	switch st {
	case SESSION_TYPE_PUSHER:
		return "pusher"
	case SESSEION_TYPE_PLAYER:
		return "player"
	}
	return "unknow"
}

type TransType int

const (
	TRANS_TYPE_TCP TransType = iota
	TRANS_TYPE_UDP
)

func (tt TransType) String() string {
	switch tt {
	case TRANS_TYPE_TCP:
		return "TCP"
	case TRANS_TYPE_UDP:
		return "UDP"
	}
	return "unknow"
}

const UDP_BUF_SIZE = 1048576

type Session struct {
	SessionLogger
	ID          string
	Server      *Server
	privateConn *RichConn
	connRW      *bufio.ReadWriter
	connWLock   sync.RWMutex
	Type        SessionType
	TransType   TransType
	Path        string
	URL         string
	SDPRaw      string
	SDPMap      map[string]*SDPInfo

	AControl string
	VControl string
	ACodec   string
	VCodec   string

	// stats info
	InBytes  int
	OutBytes int
	StartAt  time.Time
	Timeout  int

	Stoped bool

	//tcp channels
	aRTPChannel        int
	aRTPControlChannel int
	vRTPChannel        int
	vRTPControlChannel int

	Pusher    *Pusher
	Player    *Player
	UDPClient *UDPClient

	UDPSender   *net.UDPConn
	UDPBindHost string

	RTPHandles  []func(*RTPPack)
	StopHandles []func()
}

func (session *Session) AddUdpHostPort(udpHostPort string) error {
	var (
		raddr *net.UDPAddr
		laddr *net.UDPAddr
		err   error
	)
	logger := session.getLogger()

	if raddr, err = net.ResolveUDPAddr("udp4", udpHostPort); err != nil {
		logger.Printf("udp address is error [%s]", udpHostPort)
		return err
	}
	address := session.UDPBindHost
	if address != "" && strings.IndexAny(address, ":") == -1 {
		address = address + ":"
	}
	if laddr, err = net.ResolveUDPAddr("udp4", address); err != nil {
		logger.Println("error bind address:", address)
	} else {
		logger.Println("local bind address:", address)
	}
	if session.UDPSender, err = net.DialUDP("udp4", laddr, raddr); err != nil {
		logger.Printf("udp connection is error [%s]", udpHostPort)
		return err
	}
	return nil
}

func (session *Session) AddRTPHandles(f func(*RTPPack)) {
	session.RTPHandles = append(session.RTPHandles, f)
}

func (session *Session) AddStopHandles(f func()) {
	session.StopHandles = append(session.StopHandles, f)
}

func (session *Session) GetID() string {
	return session.ID
}

func (session *Session) GetPath() string {
	return session.Path
}

func (session *Session) GetConn() *RichConn {
	return session.privateConn
}

func (session *Session) String() string {
	if session.privateConn != nil {
		return fmt.Sprintf("session[%v][%v][%s][%s][%s]", session.Type, session.TransType, session.Path, session.ID, session.privateConn.RemoteAddr().String())
	} else {
		return fmt.Sprintf("session[%v][%v][%s][%s]", session.Type, session.TransType, session.Path, session.ID)
	}
}

func NewNoneConnSession(server *Server) *Session {
	session := &Session{
		ID:          shortid.MustGenerate(),
		Server:      server,
		StartAt:     time.Now(),
		Timeout:     utils.Conf().Section("rtsp").Key("timeout").MustInt(0),
		UDPBindHost: utils.Conf().Section("udp").Key("bind_host").MustString(""),

		RTPHandles:  make([]func(*RTPPack), 0),
		StopHandles: make([]func(), 0),
	}
	session.innerLogger = log.New(os.Stdout, fmt.Sprintf("[%s]*", session.ID), log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	if !utils.Debug {
		session.innerLogger.SetOutput(utils.GetLogWriter())
	}
	return session
}

func NewSession(server *Server, conn *net.TCPConn) *Session {
	networkBuffer := utils.Conf().Section("rtsp").Key("network_buffer").MustInt(204800)
	timeoutMillis := utils.Conf().Section("rtsp").Key("timeout").MustInt(0)
	timeoutTCPConn := &RichConn{conn, time.Duration(timeoutMillis) * time.Millisecond}
	session := &Session{
		ID:          shortid.MustGenerate(),
		Server:      server,
		privateConn: timeoutTCPConn,
		connRW:      bufio.NewReadWriter(bufio.NewReaderSize(timeoutTCPConn, networkBuffer), bufio.NewWriterSize(timeoutTCPConn, networkBuffer)),
		StartAt:     time.Now(),
		Timeout:     utils.Conf().Section("rtsp").Key("timeout").MustInt(0),
		UDPBindHost: utils.Conf().Section("udp").Key("bind_host").MustString(""),

		RTPHandles:  make([]func(*RTPPack), 0),
		StopHandles: make([]func(), 0),
	}
	session.innerLogger = log.New(os.Stdout, fmt.Sprintf("[%s] ", session.ID), log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	if !utils.Debug {
		session.innerLogger.SetOutput(utils.GetLogWriter())
	}
	return session
}

func (session *Session) Stop() {
	logger := session.getLogger()
	logger.Println("Session Stop()")
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
	if session.UDPSender != nil {
		session.UDPSender.Close()
		session.UDPSender = nil
	}
}

func (session *Session) Start() {
	defer session.Stop()
	buf1 := make([]byte, 1)
	buf2 := make([]byte, 2)
	logger := session.getLogger()
	for !session.Stoped {
		if _, err := io.ReadFull(session.connRW, buf1); err != nil {
			logger.Println(session, err)
			return
		}
		if buf1[0] == 0x24 { //rtp data
			if _, err := io.ReadFull(session.connRW, buf1); err != nil {
				logger.Println(err)
				return
			}
			if _, err := io.ReadFull(session.connRW, buf2); err != nil {
				logger.Println(err)
				return
			}
			channel := int(buf1[0])
			rtpLen := int(binary.BigEndian.Uint16(buf2))
			rtpBytes := make([]byte, rtpLen)
			if _, err := io.ReadFull(session.connRW, rtpBytes); err != nil {
				logger.Println(err)
				return
			}
			rtpBuf := bytes.NewBuffer(rtpBytes)
			var pack *RTPPack
			switch channel {
			case session.aRTPChannel:
				pack = &RTPPack{
					Type:   RTP_TYPE_AUDIO,
					Buffer: rtpBuf,
				}
			case session.aRTPControlChannel:
				pack = &RTPPack{
					Type:   RTP_TYPE_AUDIOCONTROL,
					Buffer: rtpBuf,
				}
			case session.vRTPChannel:
				pack = &RTPPack{
					Type:   RTP_TYPE_VIDEO,
					Buffer: rtpBuf,
				}
			case session.vRTPControlChannel:
				pack = &RTPPack{
					Type:   RTP_TYPE_VIDEOCONTROL,
					Buffer: rtpBuf,
				}
			default:
				logger.Printf("unknow rtp pack type, %v", pack.Type)
				continue
			}
			if pack == nil {
				logger.Printf("session tcp got nil rtp pack")
				continue
			}
			session.InBytes += rtpLen + 4
			for _, h := range session.RTPHandles {
				h(pack)
			}
		} else { // rtsp cmd
			reqBuf := bytes.NewBuffer(nil)
			reqBuf.Write(buf1)
			for !session.Stoped {
				if line, isPrefix, err := session.connRW.ReadLine(); err != nil {
					logger.Println(err)
					return
				} else {
					reqBuf.Write(line)
					if !isPrefix {
						reqBuf.WriteString("\r\n")
					}
					if len(line) == 0 {
						req := NewRequest(reqBuf.String())
						if req == nil {
							break
						}
						session.InBytes += reqBuf.Len()
						contentLen := req.GetContentLength()
						session.InBytes += contentLen
						if contentLen > 0 {
							bodyBuf := make([]byte, contentLen)
							if n, err := io.ReadFull(session.connRW, bodyBuf); err != nil {
								logger.Println(err)
								return
							} else if n != contentLen {
								logger.Printf("read rtsp request body failed, expect size[%d], got size[%d]", contentLen, n)
								return
							}
							req.Body = string(bodyBuf)
						}
						session.handleRequest(req)
						break
					}
				}
			}
		}
	}
}

func (session *Session) handleRequest(req *Request) {
	//if session.Timeout > 0 {
	//	session.privateConn.SetDeadline(time.Now().Add(time.Duration(session.Timeout) * time.Second))
	//}
	logger := session.getLogger()
	logger.Println("<<<", req)
	res := NewResponse(200, "OK", req.Header["CSeq"], session.ID, "")
	defer func() {
		if p := recover(); p != nil {
			res.StatusCode = 500
			res.Status = fmt.Sprintf("Inner Server Error, %v", p)
		}
		logger.Println(">>>", res)
		outBytes := []byte(res.String())
		session.connWLock.Lock()
		session.connRW.Write(outBytes)
		session.connRW.Flush()
		session.connWLock.Unlock()
		session.OutBytes += len(outBytes)
		switch req.Method {
		case "PLAY", "RECORD":
			switch session.Type {
			case SESSEION_TYPE_PLAYER:
				session.Pusher.AddPlayer(session.Player)
			case SESSION_TYPE_PUSHER:
				session.Server.AddPusher(session.Pusher)
			}
		case "TEARDOWN":
			session.Stop()
		}
	}()
	switch req.Method {
	case "OPTIONS":
		res.Header["Public"] = "DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE, OPTIONS, ANNOUNCE, RECORD"
	case "ANNOUNCE":
		session.Type = SESSION_TYPE_PUSHER
		session.URL = req.URL

		url, err := url.Parse(req.URL)
		if err != nil {
			res.StatusCode = 500
			res.Status = "Invalid URL"
			return
		}
		session.Path = url.Path

		session.SDPRaw = req.Body
		session.SDPMap = ParseSDP(req.Body)
		sdp, ok := session.SDPMap["audio"]
		if ok {
			session.AControl = sdp.Control
			session.ACodec = sdp.Codec
			logger.Printf("audio codec[%s]\n", session.ACodec)
		}
		sdp, ok = session.SDPMap["video"]
		if ok {
			session.VControl = sdp.Control
			session.VCodec = sdp.Codec
			logger.Printf("video codec[%s]\n", session.VCodec)
		}
		session.Pusher = NewPusher(session)
		if session.Server.GetPusher(session.Path) == nil {
			session.Server.AddPusher(session.Pusher)
		} else {
			res.StatusCode = 406
			res.Status = "Not Acceptable"
			return
		}
	case "DESCRIBE":
		session.Type = SESSEION_TYPE_PLAYER
		session.URL = req.URL

		url, err := url.Parse(req.URL)
		if err != nil {
			res.StatusCode = 500
			res.Status = "Invalid URL"
			return
		}
		session.Path = url.Path
		pusher := session.Server.GetPusher(session.Path)
		if pusher == nil {
			res.StatusCode = 404
			res.Status = "NOT FOUND"
			return
		}
		session.Player = NewPlayer(session, pusher)
		session.Pusher = pusher
		session.AControl = pusher.GetAControl()
		session.VControl = pusher.GetVControl()
		session.ACodec = pusher.GetACodec()
		session.VCodec = pusher.GetVCodec()
		session.privateConn.timeout = 0
		res.SetBody(session.Pusher.GetSDPRaw())
	case "SETUP":
		ts := req.Header["Transport"]
		control := req.URL[strings.LastIndex(req.URL, "/")+1:]
		mtcp := regexp.MustCompile("interleaved=(\\d+)(-(\\d+))?")
		mudp := regexp.MustCompile("client_port=(\\d+)(-(\\d+))?")

		if tcpMatchs := mtcp.FindStringSubmatch(ts); tcpMatchs != nil {
			session.TransType = TRANS_TYPE_TCP
			if control == session.AControl {
				session.aRTPChannel, _ = strconv.Atoi(tcpMatchs[1])
				session.aRTPControlChannel, _ = strconv.Atoi(tcpMatchs[3])
			} else if control == session.VControl {
				session.vRTPChannel, _ = strconv.Atoi(tcpMatchs[1])
				session.vRTPControlChannel, _ = strconv.Atoi(tcpMatchs[3])
			}
		} else if udpMatchs := mudp.FindStringSubmatch(ts); udpMatchs != nil {
			session.TransType = TRANS_TYPE_UDP
			// no need for tcp timeout.
			session.privateConn.timeout = 0
			if session.UDPClient == nil {
				session.UDPClient = &UDPClient{
					Session: session,
				}
			}
			if session.Type == SESSION_TYPE_PUSHER && session.Pusher.GetUDPServer() == nil {
				u := &UDPServer{
					Session: session,
				}
				session.Pusher.SetUDPServer(u)
			}
			if control == session.AControl {
				session.UDPClient.APort, _ = strconv.Atoi(udpMatchs[1])
				session.UDPClient.AControlPort, _ = strconv.Atoi(udpMatchs[3])
				if err := session.UDPClient.SetupAudio(); err != nil {
					res.StatusCode = 500
					res.Status = fmt.Sprintf("udp client setup audio error, %v", err)
					return
				}

				if session.Type == SESSION_TYPE_PUSHER {
					if err := session.Pusher.GetUDPServer().SetupAudio(); err != nil {
						res.StatusCode = 500
						res.Status = fmt.Sprintf("udp server setup audio error, %v", err)
						return
					}
					tss := strings.Split(ts, ";")
					idx := -1
					for i, val := range tss {
						if val == udpMatchs[0] {
							idx = i
						}
					}
					tail := append([]string{}, tss[idx+1:]...)
					tss = append(tss[:idx+1], fmt.Sprintf("server_port=%d-%d", session.Pusher.GetUDPServer().APort, session.Pusher.GetUDPServer().AControlPort))
					tss = append(tss, tail...)
					ts = strings.Join(tss, ";")
				}
			} else if control == session.VControl {
				session.UDPClient.VPort, _ = strconv.Atoi(udpMatchs[1])
				session.UDPClient.VControlPort, _ = strconv.Atoi(udpMatchs[3])
				if err := session.UDPClient.SetupVideo(); err != nil {
					res.StatusCode = 500
					res.Status = fmt.Sprintf("udp client setup video error, %v", err)
					return
				}

				if session.Type == SESSION_TYPE_PUSHER {
					if err := session.Pusher.GetUDPServer().SetupVideo(); err != nil {
						res.StatusCode = 500
						res.Status = fmt.Sprintf("udp server setup video error, %v", err)
						return
					}
					tss := strings.Split(ts, ";")
					idx := -1
					for i, val := range tss {
						if val == udpMatchs[0] {
							idx = i
						}
					}
					tail := append([]string{}, tss[idx+1:]...)
					tss = append(tss[:idx+1], fmt.Sprintf("server_port=%d-%d", session.Pusher.GetUDPServer().VPort, session.Pusher.GetUDPServer().VControlPort))
					tss = append(tss, tail...)
					ts = strings.Join(tss, ";")
				}
			}
		}
		res.Header["Transport"] = ts
	case "PLAY":
		res.Header["Range"] = req.Header["Range"]
	case "RECORD":
	}
}

func (session *Session) SessionSendRTP(pack *RTPPack) (err error) {
	if pack == nil {
		err = fmt.Errorf("player send rtp got nil pack")
		return
	}
	if session.TransType == TRANS_TYPE_UDP {
		if session.UDPClient == nil {
			err = fmt.Errorf("player use udp transport but udp client not found")
			return
		}
		err = session.UDPClient.SendRTP(pack)
		return
	}
	switch pack.Type {
	case RTP_TYPE_AUDIO:
		bufChannel := make([]byte, 2)
		bufChannel[0] = 0x24
		bufChannel[1] = byte(session.aRTPChannel)
		session.connWLock.Lock()
		session.connRW.Write(bufChannel)
		bufLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bufLen, uint16(pack.Buffer.Len()))
		session.connRW.Write(bufLen)
		session.connRW.Write(pack.Buffer.Bytes())
		session.connRW.Flush()
		session.connWLock.Unlock()
		session.OutBytes += pack.Buffer.Len() + 4
	case RTP_TYPE_AUDIOCONTROL:
		bufChannel := make([]byte, 2)
		bufChannel[0] = 0x24
		bufChannel[1] = byte(session.aRTPControlChannel)
		session.connWLock.Lock()
		session.connRW.Write(bufChannel)
		bufLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bufLen, uint16(pack.Buffer.Len()))
		session.connRW.Write(bufLen)
		session.connRW.Write(pack.Buffer.Bytes())
		session.connRW.Flush()
		session.connWLock.Unlock()
		session.OutBytes += pack.Buffer.Len() + 4
	case RTP_TYPE_VIDEO:
		bufChannel := make([]byte, 2)
		bufChannel[0] = 0x24
		bufChannel[1] = byte(session.vRTPChannel)
		session.connWLock.Lock()
		session.connRW.Write(bufChannel)
		bufLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bufLen, uint16(pack.Buffer.Len()))
		session.connRW.Write(bufLen)
		session.connRW.Write(pack.Buffer.Bytes())
		session.connRW.Flush()
		session.connWLock.Unlock()
		session.OutBytes += pack.Buffer.Len() + 4
	case RTP_TYPE_VIDEOCONTROL:
		bufChannel := make([]byte, 2)
		bufChannel[0] = 0x24
		bufChannel[1] = byte(session.vRTPControlChannel)
		session.connWLock.Lock()
		session.connRW.Write(bufChannel)
		bufLen := make([]byte, 2)
		binary.BigEndian.PutUint16(bufLen, uint16(pack.Buffer.Len()))
		session.connRW.Write(bufLen)
		session.connRW.Write(pack.Buffer.Bytes())
		session.connRW.Flush()
		session.connWLock.Unlock()
		session.OutBytes += pack.Buffer.Len() + 4
	default:
		err = fmt.Errorf("session tcp send rtp got unkown pack type[%v]", pack.Type)
	}
	return
}
