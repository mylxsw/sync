package storage

import (
	"github.com/siddontang/ledisdb/ledis"
)

// QueueFactory 队列工厂接口
type QueueFactory interface {
	// Queue 获取队列实例
	Queue(name string) Queue
}

// queueFactory 队列工厂实现
type queueFactory struct {
	db *ledis.DB
}

// NewQueueFactory 创建一个队列工厂实例
func NewQueueFactory(db *ledis.DB) QueueFactory {
	return &queueFactory{db: db}
}

// Queue 获取一个队列
func (factory *queueFactory) Queue(name string) Queue {
	return NewLedisQueue(factory.db, name)
}
