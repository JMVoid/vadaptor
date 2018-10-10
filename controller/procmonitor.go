package controller

import (
	"os/exec"
	log "github.com/Sirupsen/logrus"
	"bufio"
	"io"
	"time"
	"os"
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
		str, _, err := rd.ReadLine()
		if err != nil {
			log.Println("end of ReadLine")
			return
		}
		log.Printf("V2ray core standard out: %s", str)

	}
}

//func (m *Monitor) runV2ray(ctx context.Context) {
func (m *Monitor) runV2ray() {
	//m.cmd = exec.CommandContext(ctx, "./v2ray/v2ray", "-config", "./v2ray/v2ray.json")
	//m.cmd = exec.CommandContext(ctx, m.V2rayPath + "v2ray", "-config", m.V2rayPath + m.V2rayCfg)
	m.cmd = exec.Command(m.V2rayPath + "v2ray", "-config", m.V2rayPath + m.V2rayCfg)
	stdout, err := m.cmd.StdoutPipe()
	if err != nil {
		log.Panicf("fail to command standard out pipe %v", err)
	}
	stderr, err := m.cmd.StderrPipe()

	if err != nil {
		log.Panicf("fail to command standard error pipe %v", err)
	}

	err = m.cmd.Start()
	if err != nil {
		log.Panicf("fail to start program: %v", err)
	}

	go pipePrint(stdout)
	go pipePrint(stderr)

	err = m.cmd.Wait()
	log.Println("end of cmd.Wait")

	if err != nil {
		log.Fatalf("fail to wait%v\n", err)
	}
}

//func (m *Monitor) ProcLooper(ctx context.Context, ch chan os.Signal) {
func (m *Monitor) ProcLooper(ch chan os.Signal) {

	go m.runV2ray()
loop:
	for {
		if m.cmd != nil {
			if m.cmd.ProcessState != nil {
				log.Infoln("restarting V2ray program")
				go m.runV2ray()
			} else {
				log.Debugln("exec command process status is null, v2ray is living")
			}
		} else {
			log.Debugln("exec command is null, start v2ray")
		}
		select {
		case <-time.After(time.Duration(LOOPINTERVAL) * time.Second):
			continue
		case <-ch:
			//log.Errorln("get signal and interrupt process")
			m.cmd.Process.Signal(os.Interrupt)
			break loop
		}
	}
}
