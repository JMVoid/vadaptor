package controller

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/JMVoid/vadaptor/mysql"
	"github.com/JMVoid/vadaptor/pb"
	"github.com/JMVoid/vadaptor/utils"
	"time"
	"os"
)

//

// config
// mysql server string
// circle time
//

const cfgFile = "./app.json"

type AppConfig struct {
	DbCfg string `json:"db_config"`

	V2rayAddr  string `json:"v2ray_addr"`
	InboundTag string `json:"inbound_tag"`

	CycleSecond uint32 `json:"cycle_second"`

	NodeId uint32 `json:"node_id"`
}

type Manager struct {
	V2Inst *V2Controller
	MyDb   *mysql.DbClient

	Cfg *AppConfig

	BootTime int64

	localRepo  *pb.UserRepo
	remoteRepo *pb.UserRepo
}

func NewManager() *Manager {

	appCfg := new(AppConfig)

	manager := new(Manager)

	manager.BootTime = time.Now().Unix()

	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Fatalf("fail to read app config file with error: %v\n", err)
	}

	if err = json.Unmarshal(data, appCfg); err != nil {
		log.Fatalf("fail to parse app config file wiht error: %v\n", err)
	}
	log.Println("load app config completed")

	mydb := mysql.NewDb(appCfg.DbCfg)

	v2Controller, err := NewV2Controller(appCfg.V2rayAddr, appCfg.InboundTag)
	if err != nil {
		log.Fatalf("fail to init V2Controller with error: %v\n", err)
	}

	manager.MyDb = mydb
	manager.V2Inst = v2Controller
	manager.Cfg = appCfg
	return manager
}

func (m *Manager) initNewUsers() {
	for k, v := range m.localRepo.Usermap {
		err := m.V2Inst.AddUser(*v)
		if err != nil {
			log.Errorf("error on init new user [%s] with %v\n", k, err)
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
		log.Debugf("user %s was exist in local db, skip it", k)
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
		log.Debugf("user %s was exist in local db, skip it", k)
	}
}

const DatFile = ".adaptor.dat"

func (m *Manager) Startup() {
	var err error
	if m.localRepo, err = utils.ReadRepo(DatFile); err != nil {
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

		m.remoteRepo, err = m.MyDb.PullUser(m.Cfg.NodeId)
		if err != nil || m.remoteRepo == nil {
			log.Errorf("error on pull users from remote db, %v\n")

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
		case <-time.After(time.Duration(m.Cfg.CycleSecond) * time.Second):
			log.Debugln("An cycle is completed")
			continue
		case <-ch:
			break loop
		}

	}
}

func (m *Manager) pushTransfer() {

	for _, v := range m.localRepo.Usermap {
		err := m.V2Inst.GetTraffic(v, true)
		if err != nil {
			log.Error(err)
		}
	}
	err := m.MyDb.PushUserTransfer(m.localRepo)
	if err != nil {
		log.Errorf("Push error %v\n", err)
	}

	// clean the localRepo user upIncr downIncr
	for _, v := range m.localRepo.Usermap {
		v.UpIncr = 0
		v.DownIncr = 0
	}
}

func (m *Manager) pushNodeStatus() {
	upTime := time.Now().Unix() - m.BootTime
	loadAvg := utils.GetLoadAvg()
	err := m.MyDb.PushNodeStatus(m.Cfg.NodeId, upTime, loadAvg)
	if err != nil {
		log.Errorf("fail to push node status: %v\n", err)
	}
}
