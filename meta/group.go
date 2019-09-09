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
	"github.com/mylxsw/sync/utils/ding"
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
	Files []File `json:"files" yaml:"files"`

	From   string       `json:"from,omitempty" yaml:"from,omitempty"`
	Token  string       `json:"token,omitempty" yaml:"token,omitempty"`
	Rules  []Rule       `json:"rules,omitempty" yaml:"rules,omitempty"`
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

// SyncAction 文件同步前置后置任务
type SyncAction struct {
	Action string `json:"action,omitempty" yaml:"action,omitempty"`
	When   string `json:"when,omitempty" yaml:"when,omitempty"`

	// Match         string `json:"match,omitempty" yaml:"match,omitempty"`
	// Replace       string `json:"replace,omitempty" yaml:"replace,omitempty"`

	// --- command ---
	Command       string `json:"command,omitempty" yaml:"command,omitempty"`
	ParseTemplate bool   `json:"parse_template,omitempty" yaml:"parse_template,omitempty"`

	// --- dingding ---
	Body  string `json:"body,omitempty" yaml:"body,omitempty"`
	Token string `json:"token,omitempty" yaml:"token,omitempty"`
}

type SyncMatchData struct {
	FileSyncGroup FileSyncGroup
	Units         []SyncUnit
	Err           error
}

func NewSyncMatchData(grp FileSyncGroup, units []SyncUnit, err error) *SyncMatchData {
	return &SyncMatchData{FileSyncGroup: grp, Units: units, Err: err}
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
	defer func() {
		if err := recover(); err != nil {
			stage.Error(fmt.Sprintf("%s has a panic: %s", syncAction.Action, err))
		}
	}()

	switch syncAction.Action {
	case "command":
		return syncAction.commandHandler(data, stage)
	case "dingding":
		return syncAction.dingdingHandler(data, stage)
	default:
		stage.Error(fmt.Sprintf("[%s] %s", syncAction.Command, "not support"))
	}

	return nil
}

func (syncAction SyncAction) dingdingHandler(data *SyncMatchData, stage *collector.Stage) error {
	markdownBody := syncAction.Body
	body, err := ParseTemplate(syncAction.Body, data)
	if err != nil {
		stage.Error(fmt.Sprintf("parse dingding template [%s] failed: %s", syncAction.Body, err))
	} else {
		markdownBody = body
	}

	msg := ding.NewMarkdownMessage(
		fmt.Sprintf("Sync %s notification", data.FileSyncGroup.Name),
		markdownBody,
		[]string{},
	)

	dingClient := ding.NewDingding(syncAction.Token)
	if err := dingClient.Send(msg); err != nil {
		stage.Error(fmt.Sprintf("dingding send message failed: %s", err))
		return errors.Wrapf(err, "dingding send message failed")
	} else {
		stage.Info("dingding send message success")
	}

	return nil
}

func (syncAction SyncAction) commandHandler(data *SyncMatchData, stage *collector.Stage) error {
	commandStr := syncAction.Command
	if syncAction.ParseTemplate {
		cs, err := ParseTemplate(syncAction.Command, data)
		if err != nil {
			stage.Error(fmt.Sprintf("command [%s] parsed as template failed: %s", syncAction.Command, err))
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
			stage.Error(fmt.Sprintf("command [%s] %s", commandStr, msg))
		}

		return errors.Wrap(err, fmt.Sprintf("command [%s] execute failed", commandStr))
	}
	stage.Info(fmt.Sprintf("[%s] %s", commandStr, cmd.StdoutString()))
	return nil
}
