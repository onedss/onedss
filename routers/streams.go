package routers

import (
	"encoding/hex"
	"fmt"
	"github.com/onedss/onedss/rtprtcp"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/onedss/EasyGoLib/db"
	"github.com/onedss/onedss/models"

	"github.com/gin-gonic/gin"
	"github.com/onedss/onedss/lal/base"
	"github.com/onedss/onedss/rtmp"
	"github.com/onedss/onedss/rtsp"
)

/**
 * @apiDefine stream 流管理
 */

//StreamStart
/**
 * @api {get} /api/v1/stream/start 启动拉转推
 * @apiGroup stream
 * @apiName StreamStart
 * @apiParam {String} url RTSP源地址
 * @apiParam {String} [customPath] 转推时的推送PATH
 * @apiParam {String} [udpHostPort] 转推时的推送UDP的主机和端口，中间以冒号间隔
 * @apiParam {String=TCP,UDP} [transType=TCP] 拉流传输模式
 * @apiParam {Number} [idleTimeout] 拉流时的超时时间
 * @apiParam {Number} [heartbeatInterval] 拉流时的心跳间隔，毫秒为单位。如果心跳间隔不为0，那拉流时会向源地址以该间隔发送OPTION请求用来心跳保活
 * @apiSuccess (200) {String} ID	拉流的ID。后续可以通过该ID来停止拉流
 */
func (h *APIHandler) StreamStart(c *gin.Context) {
	type Form struct {
		URL               string `form:"url" binding:"required"`
		CustomPath        string `form:"customPath"`
		UdpHostPort       string `form:"udpHostPort"`
		TransType         string `form:"transType"`
		IdleTimeout       int    `form:"idleTimeout"`
		HeartbeatInterval int    `form:"heartbeatInterval"`
	}
	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("Pull to push err:%v", err)
		return
	}
	agent := fmt.Sprintf("EasyDarwinGo/%s", BuildVersion)
	if BuildDateTime != "" {
		agent = fmt.Sprintf("%s(%s)", agent, BuildDateTime)
	}
	if strings.IndexAny(form.URL, "rtmp://") == 0 {
		pullSession := rtmp.NewPullSession(func(option *rtmp.PullSessionOption) {
			option.PullTimeoutMs = 30000
			option.ReadAvTimeoutMs = 30000
		})
		err = pullSession.Pull(form.URL, func(msg base.RtmpMsg) {
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
		if err != nil {
			log.Printf("pull rtmp failed. err=%+v", err)
		}
		return
	}
	client, err := rtsp.NewRTSPClient(rtsp.GetServer(), form.URL, int64(form.HeartbeatInterval)*1000, agent)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if form.UdpHostPort != "" {
		hostPort := strings.ReplaceAll(form.UdpHostPort, ".", "_")
		form.CustomPath = strings.ReplaceAll(hostPort, ":", "_")
	}
	if form.CustomPath != "" && !strings.HasPrefix(form.CustomPath, "/") {
		form.CustomPath = "/" + form.CustomPath
	}
	client.CustomPath = form.CustomPath
	client.UdpHostPort = form.UdpHostPort
	switch strings.ToLower(form.TransType) {
	case "udp":
		client.TransType = rtsp.TRANS_TYPE_UDP
	case "tcp":
		fallthrough
	default:
		client.TransType = rtsp.TRANS_TYPE_TCP
	}

	pusher := rtsp.NewClientPusher(client)
	if rtsp.GetServer().GetPusher(pusher.Path()) != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Path %s already exists", client.Path))
		return
	}
	err = client.Start(time.Duration(form.IdleTimeout) * time.Second)
	if err != nil {
		log.Printf("Pull stream err :%v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pull stream err: %v", err))
		return
	}
	log.Printf("Pull to push %v success ", form)
	rtsp.GetServer().AddPusher(pusher)
	// save to db.
	var stream = models.Stream{
		URL:               form.URL,
		CustomPath:        form.CustomPath,
		IdleTimeout:       form.IdleTimeout,
		HeartbeatInterval: form.HeartbeatInterval,
	}
	if db.SQLite.Where(&models.Stream{URL: form.URL}).First(&models.Stream{}).RecordNotFound() {
		db.SQLite.Create(&stream)
	} else {
		db.SQLite.Save(&stream)
	}
	c.IndentedJSON(200, pusher.ID())
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

//StreamStop
/**
 * @api {get} /api/v1/stream/stop 停止推流
 * @apiGroup stream
 * @apiName StreamStop
 * @apiParam {String} id 拉流的ID
 * @apiUse simpleSuccess
 */
func (h *APIHandler) StreamStop(c *gin.Context) {
	type Form struct {
		ID string `form:"id" binding:"required"`
	}
	var form Form
	err := c.Bind(&form)
	if err != nil {
		log.Printf("stop pull to push err:%v", err)
		return
	}
	pushers := rtsp.GetServer().GetPushers()
	for _, v := range pushers {
		if v.ID() == form.ID {
			v.Stop()
			c.IndentedJSON(200, "OK")
			log.Printf("Stop %v success ", v)
			if v.RTSPClient != nil {
				var stream models.Stream
				stream.URL = v.RTSPClient.URL
				db.SQLite.Delete(stream)
			}
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pusher[%s] not found", form.ID))
}
