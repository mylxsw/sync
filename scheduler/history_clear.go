package scheduler

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/sync/config"
	"github.com/mylxsw/sync/storage"
)

// ClearJobHistory 清理任务执行历史纪录
func ClearJobHistory(conf *config.Config, historyStore storage.JobHistory) {
	log.Debugf("starting job history clear job, keep %d ...", conf.JobHistoryKeepSize)
	if err := historyStore.Keep(conf.JobHistoryKeepSize); err != nil {
		log.Errorf("clear job history failed: %s", err)
	}
}
