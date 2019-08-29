package meta

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mylxsw/go-toolkit/executor"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/protocol"
	"github.com/pkg/errors"
)

// File 一个待同步的文件
type File struct {
	Src    string       `json:"src" yaml:"src"`
	Dest   string       `json:"dest" yaml:"dest"`
	After  []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
	Before []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
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
	From  string `json:"from" yaml:"from"`
	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	Files  []File       `json:"files" yaml:"files"`
	Rules  []Rule       `json:"rules,omitempty" yaml:"rules"`
	Before []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
	After  []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
}

func (fsg *FileSyncGroup) Encode() []byte {
	rs, _ := json.Marshal(fsg)
	return rs
}

func (fsg *FileSyncGroup) Decode(data []byte) error {
	return json.Unmarshal(data, &fsg)
}

// SyncUnit 一个同步组中的一个文件同步
type SyncUnit struct {
	Files      []*protocol.File
	FileToSync File
}

// SyncAction 文件同步前置后置任务
type SyncAction struct {
	Action  string `json:"action,omitempty" yaml:"action,omitempty"`
	Match   string `json:"match,omitempty" yaml:"match,omitempty"`
	Replace string `json:"replace,omitempty" yaml:"replace,omitempty"`
	Command string `json:"command,omitempty" yaml:"command,omitempty"`
	When    string `json:"when,omitempty" yaml:"when,omitempty"`
}

func (after SyncAction) Matched(units []SyncUnit) bool {
	return after.When == ""
}

func (after SyncAction) Execute(units []SyncUnit, stage *collector.Stage) error {
	switch after.Action {
	case "command":
		args := strings.Split(after.Command, " ")
		cmd := executor.New(args[0], args[1:]...)
		if ok, err := cmd.Run(); !ok || err != nil {
			msg := cmd.StderrString()
			if msg != "" {
				stage.Error(fmt.Sprintf("[%s] %s", args[0], msg))
			}

			return errors.Wrap(err, fmt.Sprintf("command [%s] execute failed", after.Command))
		}

		stage.Info(fmt.Sprintf("[%s] %s", after.Command, cmd.StdoutString()))
	case "replace":

	default:
		stage.Error(fmt.Sprintf("[%s] %s", after.Command, "not support"))
	}

	return nil
}
