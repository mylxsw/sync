package storage

import (
	"errors"
	"time"

	pkgError "github.com/pkg/errors"
	"github.com/siddontang/ledisdb/ledis"
)

// ErrQueueTimeout 队列超时
var ErrQueueTimeout = errors.New("timeout")

// QueueStore 队列接口
type QueueStore interface {
	// Enqueue 加入队列
	Enqueue(payload []byte) error
	// Dequeue 从队列中读取
	// timeout > 0 则使用堵塞队列
	Dequeue(timeout time.Duration) ([]byte, error)
	// All 返回队列中所有任务
	All() ([][]byte, error)
}

// ledisQueueStore 基于Ledis实现的队列
type ledisQueueStore struct {
	db   *ledis.DB
	name []byte
}

// NewLedisQueueStore 创建一个 Ledis 队列
func NewLedisQueueStore(db *ledis.DB, name string) QueueStore {
	return &ledisQueueStore{db: db, name: []byte(name)}
}

// Enqueue 数据入队
func (qe *ledisQueueStore) Enqueue(payload []byte) error {
	_, err := qe.db.LPush(qe.name, payload)
	return err
}

// Dequeue 数据出队
func (qe *ledisQueueStore) Dequeue(timeout time.Duration) ([]byte, error) {
	if timeout > 0 {
		return qe.bDequeue(timeout)
	}

	return qe.db.RPop(qe.name)
}

func (qe *ledisQueueStore) bDequeue(timeout time.Duration) ([]byte, error) {
	item, err := qe.db.BRPop([][]byte{qe.name}, timeout)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrQueueTimeout
	}

	return item[1].([]interface{})[1].([]byte), nil
}

func (qe *ledisQueueStore) All() ([][]byte, error) {
	end, err := qe.db.LLen(qe.name)
	if err != nil {
		return nil, pkgError.Wrap(err, "query queue len failed: %s")
	}

	return qe.db.LRange(qe.name, 0, int32(end))
}
