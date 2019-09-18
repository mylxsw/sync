package scheduler

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/storage"
)

// ClearJobHistory 清理任务执行历史纪录
func ClearJobHistory(conf *config.Config, historyStore storage.JobHistoryStore) {
	log.Debugf("starting job history clear job, keep %d ...", conf.JobHistoryKeepSize)
	if err := historyStore.Keep(conf.JobHistoryKeepSize); err != nil {
		log.Errorf("clear job history failed: %s", err)
	}
}

// ClearErrors 清理错误日志
func ClearErrors(conf *config.Config, msgFactory storage.MessageFactory) {
	log.Debugf("starting job message history clear job, keep %d ...", 1000)
	if err := msgFactory.MessageStore(storage.MessageErrorType).Keep(1000); err != nil {
		log.Errorf("clear job message history failed: %s", err)
	}
}
