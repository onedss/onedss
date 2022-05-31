package core

import "github.com/onedss/onedss/rtsp"

type OnePusher interface {
	Path() string
	ID() string
	Start()
	Server() *rtsp.Server
}
