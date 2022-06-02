package rtsp

import "time"

type BasePusher interface {
	Start()

	GetPath() string
	GetSource() string
	GetTransType() string

	GetID() string
	GetVCodec() string
	GetACodec() string
	GetAControl() string
	GetVControl() string
	GetSDPRaw() string

	GetUDPServer() *UDPServer
	SetUDPServer(udpServer *UDPServer)
	GetServer() *Server

	GetInBytes() int
	GetOutBytes() int
	GetStartAt() time.Time

	QueueRTP(pack *RTPPack) BasePusher
	BroadcastRTP(pack *RTPPack) BasePusher
	GetPlayers() (players map[string]*Player)
	AddPlayer(player *Player) BasePusher
	RemovePlayer(player *Player) BasePusher
}
