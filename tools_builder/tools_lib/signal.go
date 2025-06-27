//go:build !windows

package tools_lib

import (
	"github.com/lwm-galactic/logger"
	"os"
	"os/signal"
	"syscall"
)

// linux环境下的信号量
func regExitSignals() {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Infof("got signal %v, exit", sig)
		os.Exit(11)
	}()
}
