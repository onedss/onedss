package rtsp

type BasePlayer interface {
	QueueRTP(pack *RTPPack) BasePlayer
	Start()
}
