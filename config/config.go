package config

import (
	"encoding/json"
	"strings"
)

type Config struct {
	FileTransferBufferSize int64
	RPCListenAddr          string
	RPCToken               string
	APIToken               string
	DB                     string
	FileSyncWorkerNum      int
	JobHistoryKeepSize     int64
	UseLocalDashboard      bool
	AllowFiles             []string
}

func (conf *Config) Serialize() string {
	rs, _ := json.Marshal(conf)
	return string(rs)
}

func (conf *Config) Allow(path string) bool {
	for _, al := range conf.AllowFiles {
		if strings.HasPrefix(path, al) {
			return true
		}
	}

	return false
}
