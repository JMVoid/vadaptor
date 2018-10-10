package main

import (
	"github.com/JMVoid/vadaptor/controller"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	vch := make(chan os.Signal, 1)
	mch := make(chan os.Signal, 1)

	//signal.Notify(vch,
	//	syscall.SIGHUP,
	//	syscall.SIGINT,
	//	syscall.SIGTERM,
	//	syscall.SIGQUIT,
	//)

	signal.Notify(mch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	manager := controller.NewManager()
	pm := new(controller.Monitor)
	pm.V2rayPath = manager.Cfg.V2ray.V2rayPath
	pm.V2rayCfg = manager.Cfg.V2ray.V2rayCfg

	//ctx, _ := context.WithCancel(context.Background())
	//
	go pm.ProcLooper(vch)

	//vch<-syscall.SIGTERM

	manager.Startup()
	manager.Update(mch)

	defer func(ch chan os.Signal) {
		ch <- syscall.SIGTERM
	}(vch)
}
