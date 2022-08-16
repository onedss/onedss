package routers

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onedss/onedss/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
 * @apiDefine event 事件
 */

type AlarmEvent struct {
	DeviceSerial string `form:"deviceSerial" binding:"required"`
	ChannelNo    string `form:"channelNo" binding:"required"`
	AlarmTime    int    `form:"alarmTime" binding:"required"`
	AlarmType    int    `form:"alarmType"`
	RecState     int    `form:"recState"`
	Jpg          string `form:"jpg"`
	Record       string `form:"record"`
}

func (h *APIHandler) AlarmPicture(c *gin.Context) {

}

func (h *APIHandler) AlarmRecord(c *gin.Context) {

}

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
	log.Println("开始处理数据...", alarmFile)
	out, err := os.Create(alarmFile)
	if err != nil {
		errorCode = 5
		return
	}
	defer out.Close()
	var json []string
	reader := bufio.NewReader(c.Request.Body)
	for {
		line, err := reader.ReadBytes(':')
		out.Write(line)
		json = append(json, string(line))
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		value, err := reader.ReadBytes(',')
		out.Write(value)
		if strings.Index(string(line), "jpg") > 0 {
			json = append(json, "\"\",")
		} else if strings.Index(string(line), "record") > 0 {
			json = append(json, "\"\",")
		} else {
			json = append(json, string(value))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	jsonFile := filepath.Join(alarmDir, filename+".json")
	file, err := os.Create(jsonFile)
	if err != nil {
		errorCode = 6
		return
	}
	defer file.Close()
	jsonText := strings.Join(json, "")
	io.WriteString(file, jsonText)
	log.Println("处理数据完成.", alarmFile)

	//n, err := io.Copy(out, c.Request.Body)
	//if err != nil {
	//	errorCode = 6
	//	return
	//}
	//log.Println("保存数据完成. 字节数:", n)
}

func readFromDisk(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("打开文件出错！", path, err)
		return false, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		reader.ReadSlice(',')
		//line, err := reader.ReadString(',')
		if err == io.EOF {
			break
		}
	}
	return false, nil
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
