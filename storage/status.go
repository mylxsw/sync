package storage

import (
	"fmt"

	"github.com/siddontang/ledisdb/ledis"
)

// StatusKeepDuration 状态保存有效期
const StatusKeepDuration int64 = 24 * 3600

// JobStatus 任务状态
type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusUnstable JobStatus = "unstable"
	JobStatusFailed   JobStatus = "failed"
	JobStatusOK       JobStatus = "ok"
)

// JobStatusStore 任务执行状态查询
type JobStatusStore interface {
	// Status 任务执行状态查询
	Status(id string) JobStatus
	// 更新任务执行状态
	Update(id string, status JobStatus) error
}

type jobStatusStore struct {
	db *ledis.DB
}

func NewJobStatusStore(db *ledis.DB) JobStatusStore {
	return &jobStatusStore{db: db}
}

func (j *jobStatusStore) Update(id string, status JobStatus) error {
	return j.db.SetEX([]byte(fmt.Sprintf("job-%s", id)), StatusKeepDuration, []byte(status))
}

func (j *jobStatusStore) Status(id string) JobStatus {
	res, err := j.db.Get([]byte(fmt.Sprintf("job-%s", id)))
	if err != nil {
		return ""
	}

	return JobStatus(res)
}
