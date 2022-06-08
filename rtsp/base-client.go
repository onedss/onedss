package rtsp

import (
	"bytes"
	"time"
)

type RTPType int

type RTPPack struct {
	Type   RTPType
	Buffer *bytes.Buffer
}

const (
	RTP_TYPE_AUDIO RTPType = iota
	RTP_TYPE_VIDEO
	RTP_TYPE_AUDIOCONTROL
	RTP_TYPE_VIDEOCONTROL
)

func (rt RTPType) String() string {
	switch rt {
	case RTP_TYPE_AUDIO:
		return "audio"
	case RTP_TYPE_VIDEO:
		return "video"
	case RTP_TYPE_AUDIOCONTROL:
		return "audio control"
	case RTP_TYPE_VIDEOCONTROL:
		return "video control"
	}
	return "unknow"
}

type BaseClient interface {
	String() string

	GetPath() string
	GetInitFlag() bool
	GetCustomPath() string
	GetURL() string
	GetSDPRaw() string

	Start() bool
	Stop()
	Init(timeout time.Duration) (err error)

	AddRTPHandles(func(*RTPPack))
	AddStopHandles(func())
}
