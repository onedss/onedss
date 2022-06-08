package rtsp

import (
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

	AddRTPHandles(func(*RTPPack))
	AddStopHandles(func())
}
