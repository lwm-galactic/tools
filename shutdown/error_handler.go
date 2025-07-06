package shutdown

type ErrorHandler func(error)

type ErrorReporter interface {
	OnError(err error)
}

func (h ErrorHandler) OnError(err error) {
	h(err)
}
