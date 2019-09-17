package action

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/go-toolkit/executor"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/meta"
	"github.com/pkg/errors"
)

type commandAction struct {
	syncAction *meta.SyncAction
	data       *SyncMatchData
	timeout    time.Duration
}

func newCommandAction(syncAction *meta.SyncAction, data *SyncMatchData, cc *container.Container) Action {
	act := commandAction{syncAction: syncAction, data: data}
	cc.MustResolve(func(conf *config.Config) {
		act.timeout = conf.CommandTimeout
	})
	return &act
}

// workDir return the work dir for units
// it will take the first unit in units as a work dir base
// if the dest file is a directory, return this directory
// otherwise return the directory of the dest file
func (act commandAction) workDir(units []meta.SyncUnit) string {
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

func (act commandAction) Execute(stage *collector.Stage) error {
	defer func() {
		if err := recover(); err != nil {
			stage.Error(fmt.Sprintf("%s has a panic: %s", act.syncAction.Action, err))
		}
	}()

	commandStr := act.syncAction.Command
	if act.syncAction.ParseTemplate {
		cs, err := meta.ParseTemplate(act.syncAction.Command, act.data)
		if err != nil {
			stage.Error(fmt.Sprintf("command [%s] parsed as template failed: %s", act.syncAction.Command, err))
		} else {
			commandStr = cs
		}
	}
	cmd := executor.New("sh", "-c", commandStr)
	cmd.Init(func(cmd *exec.Cmd) error {
		workDir := act.workDir(act.data.Units)
		if workDir != "" {
			cmd.Dir = workDir
		}

		return nil
	})

	timeout := act.timeout
	if act.syncAction.Timeout != "" {
		ts, err := time.ParseDuration(act.syncAction.Timeout)
		if err != nil {
			log.WithFields(log.Fields{
				"act":   act,
				"value": act.syncAction.Timeout,
			}).Errorf("invalid timeout value: %s", err)
		} else {
			if ts.Seconds() > 0 {
				timeout = ts
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if ok, err := cmd.Run(ctx); !ok || err != nil {
		msg := cmd.StderrString()
		if msg != "" {
			stage.Error(fmt.Sprintf("command [%s] %s", commandStr, msg))
		}

		return errors.Wrap(err, fmt.Sprintf("command [%s] execute failed", commandStr))
	}
	stage.Info(fmt.Sprintf("[%s] %s", commandStr, cmd.StdoutString()))
	return nil
}
