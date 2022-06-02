package rtsp

import "time"

type BasePusher interface {
	Start()

	GetPath() string
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
	GetTransType() string
	GetStartAt() time.Time
	GetSource() string

	QueueRTP(pack *RTPPack) BasePusher
	BroadcastRTP(pack *RTPPack) BasePusher
	GetPlayers() (players map[string]*Player)
	AddPlayer(player *Player) BasePusher
	RemovePlayer(player *Player) BasePusher
}
