package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/process"
	"log"
)

func init() {
	processes, _ := process.Processes()
	log.Println("当前进程数：", len(processes))
	//for _, p := range processes {
	//	name, _ := p.Name()
	//	status, _ := p.Status()
	//	log.Println(p.Pid, name, status, p.String())
	//}
}

func (h *APIHandler) GetProcessInfo(c *gin.Context) {

}
