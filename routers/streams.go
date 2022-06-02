package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onedss/onedss/rtsp"
	"log"
	"net/http"
)

type StreamStopForm struct {
	ID string `form:"id" binding:"required"`
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
