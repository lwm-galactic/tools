package shutdown

type ShutdownCallback interface {
	OnShutdown(managerName string) error
}

type ShutdownFunc func(managerName string) error

func (f ShutdownFunc) OnShutdown(managerName string) error {
	return f(managerName)
}
