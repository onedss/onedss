package rtmp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/onedss/lal/pkg/aac"
	"github.com/onedss/onedss/core"
	"github.com/onedss/onedss/lal/avc"
	"github.com/onedss/onedss/lal/base"
	"github.com/onedss/onedss/lal/hevc"
	"github.com/onedss/onedss/rtprtcp"
	"github.com/onedss/onedss/rtsp"
	"github.com/onedss/onedss/sdp"
	"github.com/onedss/onedss/utils"
	"log"
	"math/rand"
	"net/url"
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

	seq      uint16
	InBytes  int
	OutBytes int

	RTPHandles []func(*rtsp.RTPPack)

	pullSession *PullSession

	onSdp                   rtsp.OnSdp
	analyzeDone             bool
	msgCache                []base.RtmpMsg
	vps, sps, pps, asc, mpa []byte
	audioPt                 base.AvPacketPt
	videoPt                 base.AvPacketPt

	audioSsrc   uint32
	videoSsrc   uint32
	audioPacker *rtprtcp.RtpPacker
	videoPacker *rtprtcp.RtpPacker
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
	}
	return client, nil
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
	var timeoutMillis int
	if timeout == 0 {
		timeoutMillis = utils.Conf().Section("rtsp").Key("timeout").MustInt(30000)
		timeout = time.Duration(timeoutMillis) * time.Millisecond
	}
	client.onSdp = onSdp
	client.pullSession = NewPullSession(func(option *PullSessionOption) {
		option.PullTimeoutMs = timeoutMillis
		option.ReadAvTimeoutMs = timeoutMillis
	})
	client.InitFlag = true

	if err := client.pullSession.Pull(client.URL, client.onReadRtmpAvMsg); err != nil {
		return err
	}

	if client.SDPRaw != "" {
		onSdp(client.SDPRaw)
	} else {
		log.Printf("Wating a moment to callback onSdp().")
	}
	return nil
}

func (client *RTMPClient) onReadRtmpAvMsg(msg base.RtmpMsg) {
	var err error

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

	// ??????????????????????????????rtmp????????????????????????????????????
	// ?????????????????????????????????????????????????????????
	// ?????????????????????????????????Analyze????????????

	if !client.analyzeDone {
		client.msgCache = append(client.msgCache, msg.Clone())

		if msg.IsAvcKeySeqHeader() || msg.IsHevcKeySeqHeader() {
			if msg.IsAvcKeySeqHeader() {
				client.sps, client.pps, err = avc.ParseSpsPpsFromSeqHeader(msg.Clone().Payload)
				if err != nil {
					return
				}
			} else if msg.IsHevcKeySeqHeader() {
				client.vps, client.sps, client.pps, err = hevc.ParseVpsSpsPpsFromSeqHeader(msg.Clone().Payload)
				if err != nil {
					return
				}
			}
			client.doAnalyze()
			return
		}

		if msg.IsAacSeqHeader() {
			client.asc = msg.Clone().Payload[2:]
			client.doAnalyze()
			return
		}

		if msg.IsMp3SeqHeader() {
			client.mpa = msg.Clone().Payload[1:]
			client.doAnalyze()
			return
		}

		client.doAnalyze()
		return
	}

	// ????????????

	// ?????????????????????sdp?????????rtp?????????????????????????????????
	if msg.IsAvcKeySeqHeader() || msg.IsHevcKeySeqHeader() || msg.IsAacSeqHeader() {
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
		if client.mpa != nil {
			client.audioPt = base.AvPacketPtMpa
		}

		// ??????sdp
		sdpRaw, err := sdp.Pack(client.vps, client.sps, client.pps, client.asc, client.mpa, client.Agent)
		if err != nil {
			log.Println("sdp pack error!")
			return
		}
		client.SDPRaw = sdpRaw
		client.onSdp(sdpRaw)

		// ???????????????????????????
		for i := range client.msgCache {
			client.remux(client.msgCache[i])
		}
		client.msgCache = nil

		client.analyzeDone = true
	}
}

func (client *RTMPClient) isAnalyzeEnough() bool {
	var analyzeAvMsgSize int = len(client.msgCache)
	// ???????????????????????????
	if client.sps != nil && client.pps != nil && (client.asc != nil || client.mpa != nil) {
		log.Printf("audio and video is ok. analyzeAvMsgSize=%d", analyzeAvMsgSize)
		return true
	}

	// ???????????????????????????
	if analyzeAvMsgSize >= maxAnalyzeAvMsgSize {
		log.Printf("analyzeAvMsgSize=%d", analyzeAvMsgSize)
		return true
	}

	return false
}

func (client *RTMPClient) remux(msg base.RtmpMsg) {
	var packer *rtprtcp.RtpPacker
	var rtppkts []rtprtcp.RtpPacket
	switch msg.Header.MsgTypeId {
	case base.RtmpTypeIdAudio:
		packer = client.getAudioPacker()
		if packer != nil {
			rtppkts = packer.Pack(base.AvPacket{
				Timestamp:   msg.Header.TimestampAbs,
				PayloadType: client.audioPt,
				Payload:     msg.Payload[2:],
			})
		}
	case base.RtmpTypeIdVideo:
		packer = client.getVideoPacker()
		if packer != nil {
			rtppkts = client.getVideoPacker().Pack(base.AvPacket{
				Timestamp:   msg.Header.TimestampAbs,
				PayloadType: client.videoPt,
				Payload:     msg.Payload[5:],
			})
		}
	}
	for i := range rtppkts {
		pkt := rtppkts[i]
		rtpBuf := bytes.NewBuffer(pkt.Raw)
		var rtpPack *rtsp.RTPPack
		if pkt.Header.PacketType == (uint8)(client.audioPt) {
			//log.Println("PacketType =", pkt.Header.PacketType, "audioPt =", client.audioPt)
			rtpPack = &rtsp.RTPPack{
				Type:   rtsp.RTP_TYPE_AUDIO,
				Buffer: rtpBuf,
			}
		}
		if pkt.Header.PacketType == (uint8)(client.videoPt) {
			//log.Println("PacketType =", pkt.Header.PacketType, "videoPt =", client.videoPt)
			rtpPack = &rtsp.RTPPack{
				Type:   rtsp.RTP_TYPE_VIDEO,
				Buffer: rtpBuf,
			}
		}
		for _, h := range client.RTPHandles {
			if rtpPack != nil {
				h(rtpPack)
			}
		}
	}
	if len(rtppkts) > 0 {
		return
	}
	if msg.Header.MsgTypeId == base.RtmpTypeIdAudio {
		controlByte := msg.Payload[0]
		control := parseRtmpControl(controlByte)
		if control.PacketType != base.RtpPacketTypeMpa {
			log.Println("audio is not mpa.")
			return
		}
		pkg := base.AvPacket{
			Timestamp:   msg.Header.TimestampAbs,
			PayloadType: (base.AvPacketPt)(control.PacketType),
			Payload:     msg.Payload[1:],
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
func (r *RTMPClient) getAudioPacker() *rtprtcp.RtpPacker {
	if r.asc == nil {
		return nil
	}

	if r.audioPacker == nil {
		// TODO(chef): ssrc???????????????????????????lal?????????setup???????????????ssrc
		r.audioSsrc = rand.Uint32()

		ascCtx, err := aac.NewAscContext(r.asc)
		if err != nil {
			log.Printf("parse asc failed. err=%+v", err)
			return nil
		}
		clockRate, err := ascCtx.GetSamplingFrequency()
		if err != nil {
			log.Printf("get sampling frequency failed. err=%+v, asc=%s", err, hex.Dump(r.asc))
		}

		pp := rtprtcp.NewRtpPackerPayloadAac()
		r.audioPacker = rtprtcp.NewRtpPacker(pp, clockRate, r.audioSsrc)
	}
	return r.audioPacker
}

func (r *RTMPClient) getVideoPacker() *rtprtcp.RtpPacker {
	if r.sps == nil {
		return nil
	}
	if r.videoPacker == nil {
		r.videoSsrc = rand.Uint32()
		pp := rtprtcp.NewRtpPackerPayloadAvcHevc(r.videoPt, func(option *rtprtcp.RtpPackerPayloadAvcHevcOption) {
			option.Typ = rtprtcp.RtpPackerPayloadAvcHevcTypeAvcc
		})
		r.videoPacker = rtprtcp.NewRtpPacker(pp, 90000, r.videoSsrc)
	}
	return r.videoPacker
}

func init() {
	rand.Seed(time.Now().UnixNano())
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
