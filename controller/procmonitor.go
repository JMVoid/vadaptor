package controller

import (
	"os/exec"
	"context"
	log "github.com/Sirupsen/logrus"
	"bufio"
	"io"
	"os"
	"time"
)

const LOOPINTERVAL = 60

type Monitor struct {
	cmd *exec.Cmd
	V2rayPath string
	V2rayCfg string
}

func pipePrint(in io.ReadCloser) {
	rd := bufio.NewReader(in)

	for {
		str, isPrefix, err := rd.ReadLine()
		if err != nil {
			return
		}
		log.Printf("pipePrintOut: %s, isPrefix:%t\n", str, isPrefix)
	}
}

func (m *Monitor) runV2ray(ctx context.Context) {
	//m.cmd = exec.CommandContext(ctx, "./v2ray/v2ray", "-config", "./v2ray/v2ray.json")
	m.cmd = exec.CommandContext(ctx, m.V2rayPath + "v2ray", "-config", m.V2rayPath + m.V2rayCfg)
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("fail to command standard out pipe %v\n", err)
	}
	stderr, err := m.cmd.StderrPipe()

	if err != nil {
		log.Fatalf("fail to command standard error pipe %v\n", err)
	}

	err = m.cmd.Start()
	if err != nil {
		log.Fatalf("fail to start program: %v\n", err)
	}

	go pipePrint(stdout)
	go pipePrint(stderr)

	err = m.cmd.Wait()
	if err != nil {
		log.Fatalf("failt to wait%v\n", err)
	}
}

func (m *Monitor) ProcLooper(ctx context.Context, ch chan os.Signal) {
	go m.runV2ray(ctx)
loop:
	for {
		if m.cmd != nil {
			if m.cmd.ProcessState != nil {
				log.Info("restarting V2ray program")
				go m.runV2ray(ctx)
			} else {
				log.Debugln("exec command process status is null")
			}
		} else {
			log.Debugln("exec command is null")
		}
		select {
		case <-time.After(time.Duration(LOOPINTERVAL) * time.Second):
			continue
		case <-ch:
			break loop
		}
	}
}
