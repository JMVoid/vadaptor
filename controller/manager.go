package controller

import (
	log "github.com/Sirupsen/logrus"
	"github.com/JMVoid/vadaptor/mysql"
	"github.com/JMVoid/vadaptor/pb"
	"github.com/JMVoid/vadaptor/utils"
	"time"
	"os"
	"strings"
)

const InitTryTimes = 3
const CfgFile = "./config.ini"
const DatFile = ".vadaptor.dat"

type Manager struct {
	V2Inst *V2Controller
	MyDb   *mysql.DbClient

	Cfg *utils.AppConfig

	BootTime int64

	localRepo  *pb.UserRepo
	remoteRepo *pb.UserRepo
}

func levelMatch(level string) log.Level {
	switch level {
	case "panic":
		return log.PanicLevel
	case "fatal":
		return log.FatalLevel
	case "error":
		return log.ErrorLevel
	case "info":
		return log.InfoLevel
	case "debug":
		return log.DebugLevel
	default:
		return log.InfoLevel
	}

}

func NewManager() *Manager {

	appCfg := new(utils.AppConfig)

	appCfg.V2ray.V2rayAddr = "127.0.0.1:10085"
	appCfg.V2ray.InboundTag = "proxy"
	appCfg.V2ray.CycleSecond = 60
	appCfg.V2ray.V2rayPath = "./v2ray-core/"
	appCfg.V2ray.V2rayCfg = "v2ray.json"
	appCfg.V2ray.LogLevel = "info"

	manager := new(Manager)

	manager.BootTime = time.Now().Unix()

	utils.ReadConfig(CfgFile, appCfg)
	log.Println("load app config completed")

	mydb := mysql.NewDb(appCfg.V2ray)

	v2Controller, err := NewV2Controller(appCfg.V2ray.V2rayAddr, appCfg.V2ray.InboundTag)
	if err != nil {
		log.Panicf("fail to init V2Controller with error: %v\n", err)
	}

	manager.MyDb = mydb
	manager.V2Inst = v2Controller
	manager.Cfg = appCfg
	log.SetLevel(levelMatch(appCfg.V2ray.LogLevel))

	return manager
}

func (m *Manager) initNewUsers() {

	var tryCount uint32 = 0
	var keepGoing = false
	time.Sleep(2 * time.Second)
initLoop:
	for {
	loop:
		for _, v := range m.localRepo.Usermap {

			err := m.V2Inst.AddUser(*v)
			if err != nil {
				log.Println(err.Error())
				if strings.Contains(err.Error(), "connection refused") && tryCount <= InitTryTimes {

					tryCount++
					//if tryCount >= InitTryTimes {
					//	log.Panicln("try init user failure some time, program exit")
					//}
					time.Sleep(1 * time.Second)

					break loop
				} else {
					log.Panicln("init user failure, program exit")
				}
			} else {
				keepGoing = true
			}
		}
		if keepGoing {
			break initLoop
		}
	}
}

func (m *Manager) addNewUsers() {
	for k, v := range m.remoteRepo.Usermap {
		if lv, ok := m.localRepo.Usermap[k]; !ok {
			//add new user to v2ray cos not found in local db
			if v.Uplink+v.Downlink < v.TransferEnable && v.Enable > 0 {
				m.localRepo.Usermap[k] = v
				err := m.V2Inst.AddUser(*v)
				if err != nil {
					log.Errorf("error on add new user [%s] with %v\n", k, err)
				}
			}

		} else {
			lv.Enable = v.Enable
			lv.TransferEnable = v.TransferEnable
			lv.Uplink = v.Uplink
			lv.Downlink = v.Downlink
		}
		log.Debugf("user %s was exist in local db, skip add it", k)
	}
}

func (m *Manager) removeUsers() {
	for k, v := range m.localRepo.Usermap {
		if v.Uplink+v.UpIncr+v.Downlink+v.DownIncr > v.TransferEnable || v.Enable < 0 {
			err := m.V2Inst.RemoveUser(*v)
			if err != nil {
				log.Errorf("error on remove new user:%s with %v\n", k, err)
			}
			delete(m.localRepo.Usermap, k)
		}

		if _, ok := m.remoteRepo.Usermap[k]; !ok {
			// removed last user from v2ray cos not found in remote db
			err := m.V2Inst.RemoveUser(*v)
			if err != nil {
				log.Errorf("error on remove new user:%s with %v\n", k, err)
			}
			delete(m.localRepo.Usermap, k)
		}
		log.Debugf("user %s was exist in local db, skip remove it", k)
	}
}

func (m *Manager) Startup() {
	var err error

	m.localRepo, err = utils.ReadRepo(DatFile)
	if err != nil || m.localRepo == nil || m.localRepo.Usermap == nil {
		m.localRepo = new(pb.UserRepo)
		m.localRepo.Usermap = make(map[string]*pb.User)
		utils.WriteRepo(DatFile, m.localRepo)
		log.Println("created new DAT file")
	}

	m.initNewUsers()
}

func (m *Manager) Update(ch chan os.Signal) {
	var err error

loop:
	for {
		// push local data to remote Db
		m.pushTransfer()
		m.pushNodeStatus()

		m.remoteRepo, err = m.MyDb.PullUser(m.Cfg.V2ray.NodeId)
		if err != nil || m.remoteRepo == nil {
			log.Errorf("error on pull users from remote db, %v", err)

		} else {
			//push local user transfer to remote db
			// push node
			m.addNewUsers()
			m.removeUsers()
			if err = utils.WriteRepo(DatFile, m.localRepo); err != nil {
				log.Errorf("error write remote repository to dat file. %v\n", err)
			}
		}

		select {
		case <-time.After(time.Duration(m.Cfg.V2ray.CycleSecond) * time.Second):
			log.Debugln("An cycle check is completed")
			continue
		case <-ch:
			break loop
		}
	}
}

func (m *Manager) pushTransfer() error {

	for _, v := range m.localRepo.Usermap {
		err := m.V2Inst.GetTraffic(v, true)
		if err != nil {
			return err
			//log.Error(err)
		}
	}
	err := m.MyDb.PushUserTransfer(m.localRepo)
	if err != nil {
		return err
	}

	// clean the localRepo user upIncr downIncr
	for _, v := range m.localRepo.Usermap {
		v.UpIncr = 0
		v.DownIncr = 0
	}
	return nil
}

func (m *Manager) pushNodeStatus() error {
	upTime := time.Now().Unix() - m.BootTime
	loadAvg := utils.GetLoadAvg()
	err := m.MyDb.PushNodeStatus(m.Cfg.V2ray.NodeId, upTime, loadAvg)
	if err != nil {
		//log.Errorf("fail to push node status: %v", err)
		return err
	}
	return nil
}
