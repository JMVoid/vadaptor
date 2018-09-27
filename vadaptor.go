package main

import (
	"github.com/JMVoid/vadaptor/controller"
	"context"
	"os/signal"
	"syscall"
	"os"
)

func main() {

	vch := make(chan os.Signal, 1)
	mch := make(chan os.Signal, 1)
	signal.Notify(vch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	signal.Notify(mch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	pm := new(controller.Monitor)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go pm.ProcLooper(ctx, vch)

	manager := controller.NewManager()
	manager.Startup()
	manager.Update(mch)
}
