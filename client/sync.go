package client

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/protocol"
)

// File 一个待同步的文件
type File struct {
	Src    string       `json:"src"`
	Dest   string       `json:"dest"`
	After  []SyncAction `json:"after,omitempty"`
	Before []SyncAction `json:"before,omitempty"`
}

// Rule 规则
type Rule struct {
	Action
	Src string `json:"src,omitempty"`
}

// Action 共享动作结构
type Action struct {
	Action  string `json:"action,omitempty"`
	Match   string `json:"match,omitempty"`
	Replace string `json:"replace,omitempty"`
	Command string `json:"command,omitempty"`
}

// FileSyncGroup 文件同步组
type FileSyncGroup struct {
	Name  string `json:"name"`
	From  string `json:"from"`
	Token string `json:"token,omitempty"`

	Files  []File       `json:"files"`
	Rules  []Rule       `json:"rules,omitempty"`
	Before []SyncAction `json:"before,omitempty"`
	After  []SyncAction `json:"after,omitempty"`
}

// SyncAction 文件同步前置后置任务
type SyncAction struct {
	Action
	When string `json:"when,omitempty"`
}

// SyncUnit 一个同步组中的一个文件同步
type SyncUnit struct {
	Files      []*protocol.File
	FileToSync File
}

func (after SyncAction) Matched(units []SyncUnit) bool {
	return after.When == ""
}

func (after SyncAction) Execute(units []SyncUnit) error {
	log.Debugf("同 %d 个文件需要同步", len(units[0].Files))
	return nil
}
