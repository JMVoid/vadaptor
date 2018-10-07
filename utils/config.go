package utils

import (
	"gopkg.in/gcfg.v1"
	log "github.com/Sirupsen/logrus"
)

type AppConfig struct {
	V2ray V2ray
}

type V2ray struct {
	DbCfg string `gcfg:"dbConfig"`

	V2rayAddr  string `gcfg:"v2rayAddr"`
	InboundTag string `gcfg:"inboundTag"`

	CycleSecond uint32 `gcfg:"cycleSecond"`
	NodeId      uint32 `gcfg:"nodeId"`

	V2rayPath string `gcfg:"v2rayPath"`
	V2rayCfg  string `gcfg:"v2rayCfg"`
	LogLevel  string `gcfg:"logLevel"`
}

func ReadConfig(cfgFile string, config *AppConfig) {
	err := gcfg.ReadFileInto(config, cfgFile)
	if err != nil {
		log.Panicf("Fatal read config.ini config file with %v\n", err)
	}
}
