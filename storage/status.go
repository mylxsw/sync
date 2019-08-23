package storage

import (
	"fmt"

	"github.com/siddontang/ledisdb/ledis"
)

// StatusKeepDuration 状态保存有效期
const StatusKeepDuration int64 = 24 * 3600

// JobStatusQuery 任务执行状态查询
type JobStatusQuery interface {
	// Status 任务执行状态查询
	Status(id string) string
	// 更新任务执行状态
	Update(id string, status string) error
}

type jobStatusQuery struct {
	db *ledis.DB
}

func NewJobStatusQuery(db *ledis.DB) JobStatusQuery {
	return &jobStatusQuery{db: db}
}

func (j *jobStatusQuery) Update(id string, status string) error {
	return j.db.SetEX([]byte(fmt.Sprintf("job-%s", id)), StatusKeepDuration, []byte(status))
}

func (j *jobStatusQuery) Status(id string) string {
	res, err := j.db.Get([]byte(fmt.Sprintf("job-%s", id)))
	if err != nil {
		return ""
	}

	return string(res)
}
