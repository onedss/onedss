package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
 * @apiDefine event 事件
 */

/**
 * @api {get} /api/v1/alarm_event 接收推送事件
 * @apiGroup event
 * @apiName AlarmEvent
 * @apiSuccess (200) {String} result 返回码
 * @apiSuccess (200) {String} reason 描述信息
 */
func (h *APIHandler) AlarmEvent(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"result": 0,
		"reason": "ok",
	})
}
