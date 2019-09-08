package storage

import (
	"encoding/json"
	"time"

	"github.com/mylxsw/coll"
	"github.com/siddontang/ledisdb/ledis"
)

type JobHistoryStore interface {
	// Record 记录 JOB 执行历史
	Record(item JobHistoryItem) error
	// Recently 返回最近的 limit 条记录
	Recently(limit int) ([]JobHistoryItem, error)
	// Truncate 清空历史纪录
	Truncate() error
	// Keep 只保留指定数量的最新记录
	Keep(keepCount int64) error
}

type jobHistoryStore struct {
	db *ledis.DB
}

func NewJobHistoryStore(db *ledis.DB) JobHistoryStore {
	return &jobHistoryStore{db: db}
}

type JobHistoryItem struct {
	ID        string    `json:"id"`
	JobID     string    `json:"job_id"`
	Name      string    `json:"name"`
	Payload   []byte    `json:"payload"`
	Output    []byte    `json:"output"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (his *jobHistoryStore) Record(item JobHistoryItem) error {
	data, _ := json.Marshal(item)

	if _, err := his.db.LPush([]byte("history"), data); err != nil {
		return err
	}

	return nil
}

func (his *jobHistoryStore) Recently(limit int) ([]JobHistoryItem, error) {
	vals, err := his.db.LRange([]byte("history"), 0, int32(limit))
	if err != nil {
		return nil, err
	}

	var hiss []JobHistoryItem
	_ = coll.Map(vals, &hiss, func(val []byte) JobHistoryItem {
		var h JobHistoryItem
		_ = json.Unmarshal(val, &h)

		return h
	})

	return hiss, err
}

func (his *jobHistoryStore) Truncate() error {
	_, err := his.db.Del([]byte("history"))
	return err
}

func (his *jobHistoryStore) Keep(keepCount int64) error {
	curLen, err := his.db.LLen([]byte("history"))
	if err != nil {
		return err
	}

	if curLen <= keepCount {
		return nil
	}

	return his.db.LTrim([]byte("history"), 0, keepCount)
}
