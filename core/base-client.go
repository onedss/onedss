package core

import (
	"github.com/onedss/onedss/rtsp"
	"time"
)

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

	AddRTPHandles(func(*rtsp.RTPPack))
	AddStopHandles(func())
}
