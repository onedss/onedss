package rtsp

type BaseSession interface {
	GetID() string
	GetPath() string
	Start()
	Stop()
	AddRTPHandles(func(*RTPPack))
	AddStopHandles(func())
}
