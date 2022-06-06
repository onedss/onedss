package rtsp

type BaseClient interface {
	Start()
	Stop()
	AddRTPHandles(func(*RTPPack))
	AddStopHandles(func())
}
