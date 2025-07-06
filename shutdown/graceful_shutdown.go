package shutdown

import "sync"

type GSInterface interface {
	StartShutdown(manager ShutdownManager)
	ReportError(err error)
	AddShutdownCallback(cb ShutdownCallback)
}

type GracefulShutdown struct {
	callbacks    []ShutdownCallback
	managers     []ShutdownManager
	errorHandler ErrorHandler
}

func New() *GracefulShutdown {
	return &GracefulShutdown{
		callbacks: make([]ShutdownCallback, 0),
		managers:  make([]ShutdownManager, 0),
	}
}

func (gs *GracefulShutdown) AddShutdownManager(manager ShutdownManager) {
	gs.managers = append(gs.managers, manager)
}

func (gs *GracefulShutdown) AddShutdownCallback(cb ShutdownCallback) {
	gs.callbacks = append(gs.callbacks, cb)
}

func (gs *GracefulShutdown) SetErrorHandler(handler ErrorHandler) {
	gs.errorHandler = handler
}

func (gs *GracefulShutdown) Start() error {
	for _, manager := range gs.managers {
		if err := manager.Start(gs); err != nil {
			return err
		}
	}
	return nil
}

func (gs *GracefulShutdown) StartShutdown(manager ShutdownManager) {
	_ = manager.ShutdownStart()

	var wg sync.WaitGroup
	for _, cb := range gs.callbacks {
		wg.Add(1)
		go func(cb ShutdownCallback) {
			defer wg.Done()
			err := cb.OnShutdown(manager.GetName())
			if err != nil {
				gs.ReportError(err)
			}
		}(cb)
	}
	wg.Wait()

	_ = manager.ShutdownFinish()
}

func (gs *GracefulShutdown) ReportError(err error) {
	if gs.errorHandler != nil && err != nil {
		gs.errorHandler(err)
	}
}
