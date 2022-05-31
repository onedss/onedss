package rtsp

type BasePusher interface {
	Server() *Server
	ID() string
	Start()
	Path() string
	RebindSession(session *Session) bool
	RemovePlayer(player *Player) *Pusher
	AControl() string
	VControl() string
	ACodec() string
	VCodec() string
}
