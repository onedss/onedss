package routers

import "github.com/onedss/onedss/utils"

var (
	BuildVersion  = "v1.0"
	BuildDateTime = "2022-06-01"
	AgentName     = "OneDSS Client"
)

type PercentData struct {
	Time utils.DateTime `json:"time"`
	Used float64        `json:"使用"`
}

type DiskData struct {
	Disk  string `json:"disk"`
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

type CountData struct {
	Time  utils.DateTime `json:"time"`
	Total uint           `json:"总数"`
}
