package core

type OneServer interface {
	Start() (err error)
	Stop()
}
