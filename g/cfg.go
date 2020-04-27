package g

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/hetao19/toolslib/file"
)

type SSHConfig struct {
	Adrr       string `json:"addr"`
	User       string `json:"user"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
}

type Config struct {
	SSH   *SSHConfig        `json:"ssh"`
	Ports map[string]string `json:"ports"`
}

var (
	config *Config
	lock   sync.RWMutex
)

func GlobalConfig() *Config {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}
	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, " is not existent.")
	}

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, " fail: ", err)
	}

	var c Config
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail: ", err)
	}

	lock.Lock()
	defer lock.Unlock()
	config = &c

	log.Println("read config file:", cfg, "successfully")
}
