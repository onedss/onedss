package rtsp

type BaseClient interface {
	String() string

	GetPath() string
	GetInitFlag() bool
	GetCustomPath() string
	GetURL() string
	GetSDPRaw() string

	Start() bool
	Stop()

	AddRTPHandles(func(*RTPPack))
	AddStopHandles(func())
}
