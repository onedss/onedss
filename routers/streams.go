package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onedss/EasyGoLib/db"
	"github.com/onedss/onedss/models"
	"github.com/onedss/onedss/rtmp"
	"github.com/onedss/onedss/rtsp"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/**
 * @apiDefine stream 流管理
 */

type StreamStartForm struct {
	URL               string `form:"url" binding:"required"`
	CustomPath        string `form:"customPath"`
	UdpHostPort       string `form:"udpHostPort"`
	TransType         string `form:"transType"`
	IdleTimeout       int    `form:"idleTimeout"`
	HeartbeatInterval int    `form:"heartbeatInterval"`
}

type StreamStopForm struct {
	ID string `form:"id" binding:"required"`
}

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
	var form StreamStartForm
	err := c.Bind(&form)
	if err != nil {
		log.Printf("Pull to push err:%v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	l, err := url.Parse(form.URL)
	if err != nil {
		log.Printf("Url parse error:%v,%v", form.URL, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var client rtsp.BaseClient
	if l.Scheme == "rtsp" {
		client, err = createRTSPClient(form)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
	} else if l.Scheme == "rtmp" {
		client, err = createRtmpClient(form)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			return
		}
		err = client.Init(time.Duration(form.IdleTimeout) * time.Second)
		if err != nil {
			log.Printf("Pull stream err :%v", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pull stream err: %v", err))
			return
		}
		log.Printf("Pull to push %v success ", form)
		return
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Unknown Scheme : %s", form.URL))
		return
	}
	sessionPuller := rtsp.NewSessionPuller(rtsp.GetServer(), client)
	if rtsp.GetServer().GetPusher(sessionPuller.GetPath()) != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Path %s already exists", client.GetPath()))
		return
	}
	err = client.Init(time.Duration(form.IdleTimeout) * time.Second)
	if err != nil {
		log.Printf("Pull stream err :%v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pull stream err: %v", err))
		return
	}
	log.Printf("Pull to push %v success ", form)
	go sessionPuller.Start()
	// save to db.
	saveToDatabase(form)
	c.IndentedJSON(200, sessionPuller.GetID())
}

func createRtmpClient(form StreamStartForm) (rtsp.BaseClient, error) {
	agent := fmt.Sprintf("OneDSS Client/%s", BuildVersion)
	if BuildDateTime != "" {
		agent = fmt.Sprintf("%s(%s)", agent, BuildDateTime)
	}
	client, err := rtmp.NewRTMPClient(form.URL, int64(form.HeartbeatInterval)*1000, agent)
	if err != nil {
		return nil, err
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
	return client, nil
}

func createRTSPClient(form StreamStartForm) (rtsp.BaseClient, error) {
	agent := fmt.Sprintf("OneDSS Client/%s", BuildVersion)
	if BuildDateTime != "" {
		agent = fmt.Sprintf("%s(%s)", agent, BuildDateTime)
	}
	client, err := rtsp.NewRTSPClient(form.URL, int64(form.HeartbeatInterval)*1000, agent)
	if err != nil {
		return nil, err
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
	return client, nil
}

func saveToDatabase(form StreamStartForm) {
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
	var form StreamStopForm
	err := c.Bind(&form)
	if err != nil {
		log.Printf("stop pull to push err:%v", err)
		return
	}
	pushers := rtsp.GetServer().GetPushers()
	for _, v := range pushers {
		if v.GetID() == form.ID {
			v.Stop()
			c.IndentedJSON(200, "OK")
			log.Printf("Stop %v success ", v)
			//if v.RTSPClient != nil {
			//	var stream models.Stream
			//	stream.URL = v.RTSPClient.URL
			//	db.SQLite.Delete(stream)
			//}
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, fmt.Sprintf("Pusher[%s] not found", form.ID))
}
