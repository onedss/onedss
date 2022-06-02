package rtsp

type BasePusher interface {
	Start()
	QueueRTP(pack *RTPPack) *Pusher
	BroadcastRTP(pack *RTPPack) *Pusher
	GetPlayers() (players map[string]*Player)
	AddPlayer(player *Player) *Pusher
	RemovePlayer(player *Player) *Pusher
}
