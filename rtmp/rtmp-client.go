package rtmp

import (
	"encoding/hex"
	"fmt"
	"github.com/onedss/onedss/lal/base"
	"github.com/onedss/onedss/rtprtcp"
	"github.com/onedss/onedss/rtsp"
	"log"
	"net/url"
	"time"
)

type RTMPClient struct {
	URL         string
	Path        string
	CustomPath  string //custom path for pusher
	UdpHostPort string
	SDPRaw      string
	InitFlag    bool

	Agent string

	RTPHandles  []func(*rtsp.RTPPack)
	StopHandles []func()
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
		return
	}
	client = &RTMPClient{
		URL:   rawUrl,
		Path:  url.Path,
		Agent: agent,
	}
	return
}

func (client *RTMPClient) Start() bool {
	return true
}

func (client *RTMPClient) Stop() {

}

func (client *RTMPClient) Init(timeout time.Duration) error {
	pullSession := NewPullSession(func(option *PullSessionOption) {
		option.PullTimeoutMs = 30000
		option.ReadAvTimeoutMs = 30000
	})
	err := pullSession.Pull(client.URL, func(msg base.RtmpMsg) {
		if msg.Header.MsgTypeId == base.RtmpTypeIdMetadata {
			// noop
			return
		}
		if msg.Header.MsgTypeId == base.RtmpTypeIdAudio {
			controlByte := msg.Payload[0]
			control := parseRtmpControl(controlByte)
			pkg := base.AvPacket{
				Timestamp:   msg.Header.TimestampAbs,
				PayloadType: (base.AvPacketPt)(control.PacketType),
				Payload:     msg.Payload[1:],
			}
			payload := make([]byte, 4+len(pkg.Payload))
			copy(payload[4:], pkg.Payload)
			//timeUnix:=time.Now().UnixNano() / 1e6
			//nazalog.Println(timeUnix)
			h := rtprtcp.MakeDefaultRtpHeader()
			h.Mark = 0
			packetType := control.PacketType
			h.PacketType = packetType
			//h.Seq = r.genSeq()
			sampleRate := control.SampleRate
			channelNum := control.ChannelNum
			h.Timestamp = uint32(float64(pkg.Timestamp) * sampleRate * float64(channelNum))
			//h.Ssrc = r.audioSsrc
			//pkt := rtprtcp.MakeRtpPacket(h, payload)
			encodedStr := hex.EncodeToString(payload)
			log.Println(encodedStr)
		}
	})
	return err
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
