package rtmp

import (
	"bytes"
	"fmt"
	"github.com/onedss/onedss/lal/base"
	"github.com/onedss/onedss/rtprtcp"
	"github.com/onedss/onedss/rtsp"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

type RTMPClient struct {
	URL         string
	Path        string
	CustomPath  string //custom path for pusher
	UdpHostPort string
	SDPRaw      string
	InitFlag    bool

	Agent       string
	audioSsrc   uint32
	videoSsrc   uint32
	seq         uint16
	InBytes     int
	OutBytes    int
	isFirstPack bool

	RTPHandles  []func(*rtsp.RTPPack)
	StopHandles []func()

	pullSession *PullSession
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
	client.StopHandles = append(client.StopHandles, f)
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
	}
	return client, nil
}

func (client *RTMPClient) Start() bool {
	return true
}

func (client *RTMPClient) Stop() {
	if client.pullSession != nil {
		client.pullSession.Dispose()
	}
}

func (client *RTMPClient) Init(timeout time.Duration) error {
	client.pullSession = NewPullSession(func(option *PullSessionOption) {
		option.PullTimeoutMs = 30000
		option.ReadAvTimeoutMs = 30000
	})
	err := client.pullSession.Pull(client.URL, client.onReadRtmpAvMsg)
	client.InitFlag = true
	return err
}

func (client *RTMPClient) onReadRtmpAvMsg(msg base.RtmpMsg) {
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
	}
	if msg.Header.MsgTypeId == base.RtmpTypeIdMetadata {
		// noop
		return
	}
	if msg.Header.MsgTypeId == base.RtmpTypeIdAudio {
		payload := make([]byte, 4+len(pkg.Payload))
		copy(payload[4:], pkg.Payload)
		//timeUnix:=time.Now().UnixNano() / 1e6
		//log.Println(timeUnix)
		h := rtprtcp.MakeDefaultRtpHeader()
		h.Mark = 0
		packetType := control.PacketType
		h.PacketType = packetType
		h.Seq = client.genSeq()
		sampleRate := control.SampleRate
		channelNum := control.ChannelNum
		h.Timestamp = uint32(float64(pkg.Timestamp) * sampleRate * float64(channelNum))
		h.Ssrc = client.audioSsrc
		pkt := rtprtcp.MakeRtpPacket(h, payload)

		rtpBuf := bytes.NewBuffer(pkt.Raw)
		pack := &rtsp.RTPPack{
			Type:   rtsp.RTP_TYPE_AUDIO,
			Buffer: rtpBuf,
		}
		//client.InBytes += int(length + 4)
		for _, h := range client.RTPHandles {
			h(pack)
		}
	}
}

func (client *RTMPClient) NewSdp(pt uint8) string {
	tmpl := `v=0
o=- 0 0 IN IP4 127.0.0.1
c=IN IP4 127.0.0.1
t=0 0
m=audio 0 RTP/AVP %d
a=control:trackID=1`
	sdpStr := fmt.Sprintf(tmpl, pt)
	raw := strings.ReplaceAll(sdpStr, "\n", "\r\n")
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
