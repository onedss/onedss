package core

type OneServer interface {
	Start() (err error)
	Stop() (err error)
	GetPort() int
}
