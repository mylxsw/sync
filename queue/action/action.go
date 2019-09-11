package action

import (
	"github.com/antonmedv/expr"
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/meta"
)

// SyncMatchData is a data holder for syncAction templates
type SyncMatchData struct {
	JobID         string
	FileNeedSyncs meta.FileNeedSyncs // only available in file sync stage
	FileSyncGroup meta.FileSyncGroup
	Units         []meta.SyncUnit
	Err           error // only available in errors stage
}

// NewSyncMatchData create a new SyncMatchData
func NewSyncMatchData(jobID string, grp meta.FileSyncGroup, units []meta.SyncUnit, fileNeedSyncs meta.FileNeedSyncs, err error) *SyncMatchData {
	return &SyncMatchData{JobID: jobID, FileSyncGroup: grp, Units: units, Err: err, FileNeedSyncs: fileNeedSyncs,}
}

// Factory is a factory for creating Action
type Factory interface {
	Action(syncAction *meta.SyncAction, data *SyncMatchData) Action
}

type actionFactory struct {
	cc *container.Container
}

// NewActionFactory create a Factory
func NewActionFactory(cc *container.Container) Factory {
	return &actionFactory{cc: cc}
}

// matched return if the commandAction should be executed base on `When` option
// If `When` option is empty, we will think the expression is matched as a default behavior,
// Otherwise we parse the `When` expression
// If the `When` expression has some error, we just think the expression not match
func (fact actionFactory) matched(syncAction *meta.SyncAction, data *SyncMatchData) bool {
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

func (fact actionFactory) Action(syncAction *meta.SyncAction, data *SyncMatchData) Action {
	if !fact.matched(syncAction, data) {
		return nil
	}

	switch syncAction.Action {
	case "command":
		return newCommandAction(syncAction, data)
	case "dingding":
		return newDingdingAction(syncAction, data, fact.cc)
	}

	return nil
}

type Action interface {
	Execute(stage *collector.Stage) error
}
