//go:build !windows

package tools_lib

import (
	"git.pinquest.cn/base/log"
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
		log.Infof("got signal %v, exit", sig)
		os.Exit(11)
	}()
}
