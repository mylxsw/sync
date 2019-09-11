package action

import (
	"fmt"

	"github.com/mylxsw/container"
	"github.com/mylxsw/sync/collector"
	"github.com/mylxsw/sync/meta"
	"github.com/mylxsw/sync/storage"
	"github.com/mylxsw/sync/utils/ding"
)

type dingdingAction struct {
	syncAction *meta.SyncAction
	data       *SyncMatchData
	ddQueue    storage.QueueStore
}

func newDingdingAction(syncAction *meta.SyncAction, data *SyncMatchData, cc *container.Container) Action {
	act := dingdingAction{syncAction: syncAction, data: data,}

	cc.MustResolve(func(queueFact storage.QueueStoreFactory) {
		act.ddQueue = queueFact.Queue(storage.QueueDingding)
	})

	return &act
}

func (act dingdingAction) Execute(stage *collector.Stage) error {
	defer func() {
		if err := recover(); err != nil {
			stage.Error(fmt.Sprintf("%s has a panic: %s", act.syncAction.Action, err))
		}
	}()

	markdownBody := act.syncAction.Body
	body, err := meta.ParseTemplate(act.syncAction.Body, act.data)
	if err != nil {
		stage.Error(fmt.Sprintf("parse dingding template [%s] failed: %s", act.syncAction.Body, err))
	} else {
		markdownBody = body
	}

	msg := ding.NewMarkdownMessage(
		fmt.Sprintf("Sync %s notification", act.data.FileSyncGroup.Name),
		markdownBody,
		[]string{},
	)

	dingdingMessage := ding.DingdingMessage{
		Message: msg,
		Token:   act.syncAction.Token,
	}

	if err := act.ddQueue.Enqueue(dingdingMessage.Encode()); err != nil {
		stage.Error(fmt.Sprintf("dingding message enqueue failed: %s", err))
	} else {
		stage.Info("dingding message enqueued")
	}

	return nil
}
