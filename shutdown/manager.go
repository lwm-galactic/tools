package shutdown

type ShutdownManager interface {
	GetName() string
	Start(gs GSInterface) error
	ShutdownStart() error
	ShutdownFinish() error
}
