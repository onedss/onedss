package routers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onedss/onedss/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

/**
 * @apiDefine event 事件
 */

/**
 * @api {get} /api/v1/event/receive 接收推送事件
 * @apiGroup event
 * @apiName AlarmEvent
 * @apiSuccess (200) {String} code 返回码
 * @apiSuccess (200) {String} msg 描述信息
 */
func (h *APIHandler) AlarmEvent(c *gin.Context) {
	var contentLength int64
	errorCode := 0
	defer func() {
		if errorCode == 0 {
			c.IndentedJSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "Success",
			})
		} else {
			c.IndentedJSON(http.StatusBadRequest, gin.H{
				"code": errorCode,
				"msg":  "Failure",
			})
		}
	}()
	contentLength = c.Request.ContentLength
	if contentLength <= 0 || contentLength > 1024*1024*1024*2 {
		log.Printf("content_length error\n")
		errorCode = 1
		return
	}
	contentTypes, has_key := c.Request.Header["Content-Type"]
	if !has_key {
		log.Printf("Content-Type error\n")
		errorCode = 2
		return
	}
	if len(contentTypes) != 1 {
		log.Printf("Content-Type count error\n")
		errorCode = 3
		return
	}
	contentType := contentTypes[0]
	log.Println("Content-Type:", contentType)
	log.Println("Content-Length:", contentLength)
	filename := fmt.Sprintf("camera_%s", time.Now().Format("2006_0102_150405"))
	alarmDir := filepath.Join(utils.DataDir(), "www", "alarm", time.Now().Format("20060102"))
	flag, _ := checkDirExist(alarmDir)
	if !flag {
		err := os.MkdirAll(alarmDir, 0777)
		if err != nil {
			log.Println("Mkdir error! ", err)
			errorCode = 4
			return
		}
	}
	alarmFile := filepath.Join(alarmDir, filename)
	fmt.Println("开始处理数据...", alarmFile)
	out, err := os.Create(alarmFile)
	if err != nil {
		errorCode = 5
		return
	}
	defer out.Close()

	n, err := io.Copy(out, c.Request.Body)
	if err != nil {
		errorCode = 5
		return
	}
	//buf := make([]byte, 1024)
	//left := (int)(contentLength)
	//for {
	//	// 接收服务端信息
	//	n, err := c.Request.Body.Read(buf)
	//	if n > 0 {
	//		res := string(buf[:n])
	//		fmt.Print(len(res))
	//	}
	//	if n == 0 && err == io.EOF {
	//		break
	//	}
	//	left = left - n
	//	if left <= 0 {
	//		break
	//	}
	//	if err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//}
	fmt.Println("处理数据完成. 字节数:", n)
}

func checkDirExist(name string) (bool, error) {
	dir, err := os.Stat(name)
	if err == nil {
		if dir.IsDir() {
			return true, nil
		}
		return false, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
