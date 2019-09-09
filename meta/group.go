package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/antonmedv/expr"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/go-toolkit/executor"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/protocol"
	"github.com/pkg/errors"
)

// File 一个待同步的文件
type File struct {
	Src     string       `json:"src" yaml:"src"`
	Dest    string       `json:"dest" yaml:"dest"`
	Ignores []string     `json:"ignores,omitempty" yaml:"ignores,omitempty"`
	After   []SyncAction `json:"after,omitempty" yaml:"after,omitempty"`
	Before  []SyncAction `json:"before,omitempty" yaml:"before,omitempty"`
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
	Errors []SyncAction `json:"errors,omitempty" yaml:"errors,omitempty"`
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
	Action        string `json:"action,omitempty" yaml:"action,omitempty"`
	Match         string `json:"match,omitempty" yaml:"match,omitempty"`
	Replace       string `json:"replace,omitempty" yaml:"replace,omitempty"`
	Command       string `json:"command,omitempty" yaml:"command,omitempty"`
	When          string `json:"when,omitempty" yaml:"when,omitempty"`
	ParseTemplate bool   `json:"parse_template,omitempty" yaml:"parse_template,omitempty"`
}

type SyncMatchData struct {
	Units []SyncUnit
	Err   error
}

func NewSyncMatchData(units []SyncUnit, err error) *SyncMatchData {
	return &SyncMatchData{Units: units, Err: err}
}

// Matched return if the action should be executed base on `When` option
// If `When` option is empty, we will think the expression is matched as a default behavior,
// Otherwise we parse the `When` expression
// If the `When` expression has some error, we just think the expression not match
func (syncAction SyncAction) Matched(data *SyncMatchData) bool {
	if syncAction.When == "" {
		return true
	}

	program, err := expr.Compile(syncAction.When, expr.Env(&SyncMatchData{}), expr.AsBool())
	if err != nil {
		log.WithFields(log.Fields{
			"action": syncAction,
		}).Errorf("invalid expr, can not compile expr(%s): %s", syncAction.When, err)
		return false
	}

	rs, err := expr.Run(program, data)
	if err != nil {
		log.WithFields(log.Fields{
			"action": syncAction,
		}).Errorf("run expr failed, can not run expr(%s): %s", syncAction.When, err)
		return false
	}

	if boolRes, ok := rs.(bool); ok {
		return boolRes
	}

	log.WithFields(log.Fields{
		"action": syncAction,
	}).Errorf("invalid return value for expr (%s), not a boolean value", syncAction.When)

	return false
}

// workDir return the work dir for units
// it will take the first unit in units as a work dir base
// if the dest file is a directory, return this directory
// otherwise return the directory of the dest file
func (syncAction SyncAction) workDir(units []SyncUnit) string {
	if len(units) > 0 {
		destStat, err := os.Stat(units[0].FileToSync.Dest)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Errorf("can not get file %s stat, skip work dir setting: %s", units[0].FileToSync.Dest, err)
			}
			return ""
		}

		if destStat.IsDir() {
			return units[0].FileToSync.Dest
		} else {
			return filepath.Dir(units[0].FileToSync.Dest)
		}
	}

	return ""
}

func (syncAction SyncAction) Execute(data *SyncMatchData, stage *collector.Stage) error {
	switch syncAction.Action {
	case "command":
		commandStr := syncAction.Command
		if syncAction.ParseTemplate {
			cs, err := ParseTemplate(syncAction.Command, data)
			if err != nil {
				stage.Error(fmt.Sprintf("parse command [%s] as template failed: %s", syncAction.Command, err))
			} else {
				commandStr = cs
			}
		}

		cmd := executor.New("sh", "-c", commandStr)
		cmd.Init(func(cmd *exec.Cmd) error {
			workDir := syncAction.workDir(data.Units)
			if workDir != "" {
				cmd.Dir = workDir
			}

			return nil
		})
		if ok, err := cmd.Run(); !ok || err != nil {
			msg := cmd.StderrString()
			if msg != "" {
				stage.Error(fmt.Sprintf("[%s] %s", syncAction.Command, msg))
			}

			return errors.Wrap(err, fmt.Sprintf("command [%s] execute failed", syncAction.Command))
		}

		stage.Info(fmt.Sprintf("[%s] %s", syncAction.Command, cmd.StdoutString()))
	case "replace":

	default:
		stage.Error(fmt.Sprintf("[%s] %s", syncAction.Command, "not support"))
	}

	return nil
}
