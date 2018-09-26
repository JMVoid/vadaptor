package main

import (
	"github.com/JMVoid/v2ssadaptor/controller"
	"context"
	"os/signal"
	"syscall"
	"os"
)

func main() {

	ch := make(chan os.Signal, 1)

	pm := new(controller.Monitor)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go pm.ProcLooper(ctx, ch)

	manager := controller.NewManager()
	manager.Startup()

	signal.Notify(ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	manager.Update(ch)
}
