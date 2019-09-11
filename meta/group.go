package meta

import (
	"encoding/json"

	"github.com/mylxsw/sync/protocol"
)

// File 一个待同步的文件
type File struct {
	Src           string       `json:"src" yaml:"src"`
	Dest          string       `json:"dest" yaml:"dest"`
	Delete        bool         `json:"delete,omitempty" yaml:"delete,omitempty"`
	Ignores       []string     `json:"ignores,omitempty" yaml:"ignores,omitempty"`
	After         []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
	Before        []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
	Error         []SyncAction `json:"error,omitempty" yaml:"error,omitempty"`
	SkipWhenError bool         `json:"skip_when_error,omitempty" yaml:"skip_when_error,omitempty"`
}

// Rule 规则
type Rule struct {
	Action  string `json:"action,omitempty" yaml:"action,omitempty"`
	Match   string `json:"match,omitempty" yaml:"match,omitempty"`
	Replace string `json:"replace,omitempty" yaml:"replace,omitempty"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	Src     string `json:"src,omitempty" yaml:"src,omitempty"`
}

// FileSyncGroup 文件同步组
type FileSyncGroup struct {
	Name  string `json:"name" yaml:"name"`
	Files []File `json:"files" yaml:"files"`

	From   string       `json:"from,omitempty" yaml:"from,omitempty"`
	Token  string       `json:"token,omitempty" yaml:"token,omitempty"`
	Rules  []Rule       `json:"rules,omitempty" yaml:"rules,omitempty"`
	Before []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
	After  []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
	Error  []SyncAction `json:"error,omitempty" yaml:"error,omitempty"`
}

func (fsg *FileSyncGroup) Encode() []byte {
	rs, _ := json.Marshal(fsg)
	return rs
}

func (fsg *FileSyncGroup) Decode(data []byte) error {
	return json.Unmarshal(data, &fsg)
}

// GlobalFileSyncSetting global file sync settings
type GlobalFileSyncSetting struct {
	From   string       `json:"from,omitempty" yaml:"from,omitempty"`
	Token  string       `json:"token,omitempty" yaml:"token,omitempty"`
	Before []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
	After  []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
	Errors []SyncAction `json:"errors,omitempty" yaml:"errors,omitempty"`
}

func NewGlobalFileSyncSetting() *GlobalFileSyncSetting {
	return &GlobalFileSyncSetting{
		Before: make([]SyncAction, 0),
		After:  make([]SyncAction, 0),
		Errors: make([]SyncAction, 0),
	}
}

func (gfs *GlobalFileSyncSetting) Encode() []byte {
	rs, _ := json.Marshal(gfs)
	return rs
}

func (gfs *GlobalFileSyncSetting) Decode(data []byte) error {
	return json.Unmarshal(data, &gfs)
}

// SyncUnit 一个同步组中的一个文件同步
type SyncUnit struct {
	Files      []*protocol.File
	FileToSync File
}
