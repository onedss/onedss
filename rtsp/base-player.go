package rtsp

type BasePlayer interface {
	QueueRTP(pack *RTPPack) *Player
	Start()
}
