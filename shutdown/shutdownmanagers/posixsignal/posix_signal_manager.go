package posixsignal

import (
	"fmt"
	"github.com/lwm-galactic/tools/shutdown"

	"os"
	"os/signal"
	"syscall"
)

type PosixSignalManager struct{}

func NewPosixSignalManager() *PosixSignalManager {
	return &PosixSignalManager{}
}

func (p *PosixSignalManager) GetName() string {
	return "PosixSignalManager"
}

func (p *PosixSignalManager) Start(gs shutdown.GSInterface) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("收到关闭信号")
		gs.StartShutdown(p)
	}()
	return nil
}

func (p *PosixSignalManager) ShutdownStart() error {
	fmt.Println("开始执行关闭前准备...")
	return nil
}

func (p *PosixSignalManager) ShutdownFinish() error {
	fmt.Println("关闭完成，退出程序")
	os.Exit(0)
	return nil
}
