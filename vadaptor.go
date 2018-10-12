package main

import (
	"github.com/JMVoid/vadaptor/controller"
	"os"
	"os/signal"
	"syscall"
	"sync"
)

func main() {

	vch := make(chan os.Signal, 1)
	mch := make(chan os.Signal, 1)
	var wg sync.WaitGroup


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

	wg.Add(1)
	go pm.ProcLooper(vch, &wg)
	manager.Startup()
	manager.Update(mch)

	defer func(ch chan os.Signal) {
		ch <- syscall.SIGTERM
		wg.Wait()
	}(vch)
}
