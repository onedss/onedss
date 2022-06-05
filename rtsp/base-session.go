package rtsp

type BaseSession interface {
	GetID() string
	GetPath() string
	Start()
	Stop()
}
