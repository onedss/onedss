package rtmp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/onedss/lal/pkg/sdp"
	"github.com/onedss/onedss/core"
	"github.com/onedss/onedss/lal/base"
	"github.com/onedss/onedss/rtprtcp"
	"github.com/onedss/onedss/rtsp"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

var (
	// config
	maxAnalyzeAvMsgSize = 16
)

type RTMPClient struct {
	core.SessionLogger

	Stoped      bool
	URL         string
	Path        string
	CustomPath  string //custom path for pusher
	UdpHostPort string
	SDPRaw      string
	InitFlag    bool

	Agent string

	seq         uint16
	InBytes     int
	OutBytes    int
	isFirstPack bool

	RTPHandles []func(*rtsp.RTPPack)

	pullSession *PullSession

	onSdp              rtsp.OnSdp
	analyzeDone        bool
	msgCache           []base.RtmpMsg
	vps, sps, pps, asc []byte
	audioPt            base.AvPacketPt
	videoPt            base.AvPacketPt

	audioSsrc uint32
	videoSsrc uint32
}

func (client *RTMPClient) String() string {
	return fmt.Sprintf("client[%s]", client.URL)
}

func (client *RTMPClient) GetPath() string {
	return client.Path
}

func (client *RTMPClient) GetInitFlag() bool {
	return client.InitFlag
}

func (client *RTMPClient) GetCustomPath() string {
	return client.CustomPath
}

func (client *RTMPClient) GetURL() string {
	return client.URL
}

func (client *RTMPClient) GetSDPRaw() string {
	return client.SDPRaw
}

func (client *RTMPClient) AddRTPHandles(f func(*rtsp.RTPPack)) {
	client.RTPHandles = append(client.RTPHandles, f)
}

func (client *RTMPClient) AddStopHandles(f func()) {
	client.pullSession.AddDisposeHandles(f)
}

func NewRTMPClient(rawUrl string, sendOptionMillis int64, agent string) (client *RTMPClient, err error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	client = &RTMPClient{
		URL:       rawUrl,
		Path:      url.Path,
		Agent:     agent,
		audioSsrc: rand.Uint32(),
		videoSsrc: rand.Uint32(),
		SDPRaw:    createSDPRaw(14),
	}
	return client, nil
}

func createSDPRaw(pt uint8) string {
	if pt == 0 {
		pt = 14
	}
	tmpl := `v=0
o=- 0 0 IN IP4 127.0.0.1
c=IN IP4 127.0.0.1
t=0 0
m=audio 0 RTP/AVP %d
a=control:trackID=1`
	sdpStr := fmt.Sprintf(tmpl, pt)
	return sdpStr
}

func (client *RTMPClient) Start() bool {
	return client.InitFlag
}

func (client *RTMPClient) Stop() {
	logger := client.GetLogger()
	logger.Printf("RTMPClient Stop. [%s], Stoped = %v", client.URL, client.Stoped)
	if client.Stoped {
		return
	}
	client.Stoped = true
	if client.pullSession != nil {
		client.pullSession.Dispose()
	}
}

func (client *RTMPClient) Init(timeout time.Duration, onSdp rtsp.OnSdp) error {
	client.onSdp = onSdp
	client.pullSession = NewPullSession(func(option *PullSessionOption) {
		option.PullTimeoutMs = 30000
		option.ReadAvTimeoutMs = 30000
	})
	if err := client.pullSession.Pull(client.URL, client.onReadRtmpAvMsg); err != nil {
		return err
	}
	if client.SDPRaw != "" {
		onSdp(client.SDPRaw)
	} else {
		log.Printf("Wating a moment to callback onSdp().")
	}

	client.InitFlag = true
	return nil
}

func (client *RTMPClient) onReadRtmpAvMsg(msg base.RtmpMsg) {
	switch msg.Header.MsgTypeId {
	case base.RtmpTypeIdMetadata:
		return
	case base.RtmpTypeIdAudio:
		if len(msg.Payload) <= 2 {
			log.Printf("rtmp msg too short, ignore. header=%+v, payload=%s", msg.Header, hex.Dump(msg.Payload))
			return
		}
	case base.RtmpTypeIdVideo:
		if len(msg.Payload) <= 5 {
			log.Printf("rtmp msg too short, ignore. header=%+v, payload=%s", msg.Header, hex.Dump(msg.Payload))
			return
		}
	}

	if msg.Header.MsgTypeId == base.RtmpTypeIdMetadata {
		// noop
		return
	}
	client.remux(msg)
}

func (client *RTMPClient) doAnalyze() {
	if client.analyzeDone {
		return
	}

	if client.isAnalyzeEnough() {
		if client.sps != nil && client.pps != nil {
			if client.vps != nil {
				client.videoPt = base.AvPacketPtHevc
			} else {
				client.videoPt = base.AvPacketPtAvc
			}
		}
		if client.asc != nil {
			client.audioPt = base.AvPacketPtAac
		}

		// 回调sdp
		_, err := sdp.Pack(client.vps, client.sps, client.pps, client.asc)
		if err != nil {
			log.Println("sdp pack error!")
			return
		}
		//client.onSdp(ctx)

		// 分析阶段缓存的数据
		for i := range client.msgCache {
			client.remux(client.msgCache[i])
		}
		client.msgCache = nil

		client.analyzeDone = true
	}
}

func (client *RTMPClient) isAnalyzeEnough() bool {
	// 音视频头都收集好了
	if client.sps != nil && client.pps != nil && client.asc != nil {
		return true
	}

	// 达到分析包数阈值了
	if len(client.msgCache) >= maxAnalyzeAvMsgSize {
		return true
	}

	return false
}

func (client *RTMPClient) remux(msg base.RtmpMsg) {
	if msg.Header.MsgTypeId == base.RtmpTypeIdAudio {
		controlByte := msg.Payload[0]
		control := parseRtmpControl(controlByte)
		pkg := base.AvPacket{
			Timestamp:   msg.Header.TimestampAbs,
			PayloadType: (base.AvPacketPt)(control.PacketType),
			Payload:     msg.Payload[1:],
		}
		if !client.isFirstPack {
			client.isFirstPack = true
			client.SDPRaw = client.NewSdp(control.PacketType)
			log.Println(client.SDPRaw)
			log.Printf("%+v", control)
		}
		payload := make([]byte, 4+len(pkg.Payload))
		copy(payload[4:], pkg.Payload)
		//timeUnix:=time.Now().UnixNano() / 1e6
		//log.Println(pkg.Timestamp)
		h := rtprtcp.MakeDefaultRtpHeader()
		h.Mark = 0
		h.PacketType = control.PacketType
		h.Seq = client.genSeq()
		h.Timestamp = uint32(float64(pkg.Timestamp) * control.SampleRate * float64(control.ChannelNum))
		h.Ssrc = client.audioSsrc
		pkt := rtprtcp.MakeRtpPacket(h, payload)
		//log.Println(h.Timestamp, pkg.Timestamp, control.SampleRate, control.ChannelNum)

		rtpBuf := bytes.NewBuffer(pkt.Raw)
		pack := &rtsp.RTPPack{
			Type:   rtsp.RTP_TYPE_AUDIO,
			Buffer: rtpBuf,
		}
		for _, h := range client.RTPHandles {
			h(pack)
		}
	}
}

func (client *RTMPClient) NewSdp(pt uint8) string {
	sdpStr := createSDPRaw(pt)
	temp := strings.ReplaceAll(sdpStr, "\r\n", "\n")
	raw := strings.ReplaceAll(temp, "\n", "\r\n")
	return raw
}

func (client *RTMPClient) genSeq() (ret uint16) {
	client.seq++
	ret = client.seq
	return
}

func parseRtmpControl(control byte) rtprtcp.RtpControl {
	format := control >> 4 & 0xF
	sampleRate := control >> 2 & 0x3
	sampleSize := control >> 1 & 0x1
	channelNum := control & 0x1
	rtmpBodyControl := rtprtcp.MakeDefaultRtpControl()
	rtmpBodyControl.Format = format
	switch format {
	case base.RtmpControlMP3:
		rtmpBodyControl.PacketType = uint8(base.RtpPacketTypeMpa)
	case base.RtmpControlAAC:
		rtmpBodyControl.PacketType = uint8(base.RtpPacketTypeAac)
	default:
		rtmpBodyControl.PacketType = uint8(base.RtpPacketTypeMpa)
	}
	if sampleRate == 3 {
		rtmpBodyControl.SampleRate = 44.1
	}
	if sampleSize == 1 {
		rtmpBodyControl.SampleSize = 16
	}
	if channelNum == 1 {
		rtmpBodyControl.ChannelNum = 2
	}
	return rtmpBodyControl
}
