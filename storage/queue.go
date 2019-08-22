package storage

import (
	"errors"
	"time"

	"github.com/siddontang/ledisdb/ledis"
)

// ErrQueueTimeout 队列超时
var ErrQueueTimeout = errors.New("timeout")

// Queue 队列接口
type Queue interface {
	// Enqueue 加入队列
	Enqueue(payload []byte) error
	// Dequeue 从队列中读取
	// timeout > 0 则使用堵塞队列
	Dequeue(timeout time.Duration) ([]byte, error)
}

// ledisQueue 基于Ledis实现的队列
type ledisQueue struct {
	db   *ledis.DB
	name []byte
}

// NewLedisQueue 创建一个 Ledis 队列
func NewLedisQueue(db *ledis.DB, name string) Queue {
	return &ledisQueue{db: db, name: []byte(name)}
}

// Enqueue 数据入队
func (qe *ledisQueue) Enqueue(payload []byte) error {
	_, err := qe.db.LPush(qe.name, payload)
	return err
}

// Dequeue 数据出队
func (qe *ledisQueue) Dequeue(timeout time.Duration) ([]byte, error) {
	if timeout > 0 {
		return qe.bDequeue(timeout)
	}

	return qe.db.RPop(qe.name)
}

func (qe *ledisQueue) bDequeue(timeout time.Duration) ([]byte, error) {
	item, err := qe.db.BRPop([][]byte{qe.name}, timeout)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, ErrQueueTimeout
	}

	return item[1].([]interface{})[1].([]byte), nil
}
