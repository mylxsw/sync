package storage

import (
	"encoding/json"
	"fmt"
	"time"

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
	// LastStatus return last changed status for a sync definition
	LastStatus(name string) (JobStatus, time.Time)
	// 更新任务执行状态
	Update(id string, name string, status JobStatus) error
}

type jobStatusStore struct {
	db *ledis.DB
}

func NewJobStatusStore(db *ledis.DB) JobStatusStore {
	return &jobStatusStore{db: db}
}

type jobLastSyncStatus struct {
	Status        JobStatus `json:"status"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func (j *jobStatusStore) Update(id string, name string, status JobStatus) error {
	err := j.db.SetEX([]byte(fmt.Sprintf("job-%s", id)), StatusKeepDuration, []byte(status))
	if err != nil {
		return err
	}

	data, _ := json.Marshal(jobLastSyncStatus{
		Status:        status,
		LastUpdatedAt: time.Now(),
	})

	return j.db.Set([]byte(fmt.Sprintf("sync-status-%s", name)), data)
}

func (j *jobStatusStore) Status(id string) JobStatus {
	res, err := j.db.Get([]byte(fmt.Sprintf("job-%s", id)))
	if err != nil {
		return ""
	}

	return JobStatus(res)
}

func (j *jobStatusStore) LastStatus(name string) (JobStatus, time.Time) {
	data, err := j.db.Get([]byte(fmt.Sprintf("sync-status-%s", name)))
	if err != nil {
		return "", time.Time{}
	}
	var js jobLastSyncStatus
	_ = json.Unmarshal(data, &js)

	return js.Status, js.LastUpdatedAt
}
