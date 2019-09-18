package storage

import (
	"github.com/mylxsw/coll"
	"github.com/siddontang/ledisdb/ledis"
)

const (
	MessageErrorType = "errors"
)

type MessageFactory interface {
	MessageStore(name string) MessageStore
}

type messageFactory struct {
	db *ledis.DB
}

func NewMessageFactory(db *ledis.DB) MessageFactory {
	return &messageFactory{db: db}
}

func (m messageFactory) MessageStore(name string) MessageStore {
	return newMessageStore(m.db, name)
}

type MessageStore interface {
	// Record 记录
	Record(item string) error
	// Recently 返回最近的 limit 条记录
	Recently(limit int) ([]string, error)
	// Truncate 清空历史纪录
	Truncate() error
	// Keep 只保留指定数量的最新记录
	Keep(keepCount int64) error
}

type messageStore struct {
	db   *ledis.DB
	name []byte
}

func newMessageStore(db *ledis.DB, name string) MessageStore {
	return &messageStore{db: db, name: []byte(name),}
}

func (ms *messageStore) Record(item string) error {
	if _, err := ms.db.LPush(ms.name, []byte(item)); err != nil {
		return err
	}

	return nil
}

func (ms *messageStore) Recently(limit int) ([]string, error) {
	vals, err := ms.db.LRange(ms.name, 0, int32(limit))
	if err != nil {
		return nil, err
	}

	var hiss []string
	_ = coll.Map(vals, &hiss, func(val []byte) string {
		return string(val)
	})

	return hiss, err
}

func (ms *messageStore) Truncate() error {
	_, err := ms.db.Del(ms.name)
	return err
}

func (ms *messageStore) Keep(keepCount int64) error {
	curLen, err := ms.db.LLen(ms.name)
	if err != nil {
		return err
	}

	if curLen <= keepCount {
		return nil
	}

	return ms.db.LTrim(ms.name, 0, keepCount)
}
