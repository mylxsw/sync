package storage

import (
	"github.com/siddontang/ledisdb/ledis"
)

// QueueStoreFactory 队列工厂接口
type QueueStoreFactory interface {
	// QueueStore 获取队列实例
	Queue(name string) QueueStore
}

// queueStoreFactory 队列工厂实现
type queueStoreFactory struct {
	db *ledis.DB
}

// NewQueueStoreFactory 创建一个队列工厂实例
func NewQueueStoreFactory(db *ledis.DB) QueueStoreFactory {
	return &queueStoreFactory{db: db}
}

// QueueStore 获取一个队列
func (factory *queueStoreFactory) Queue(name string) QueueStore {
	return NewLedisQueueStore(factory.db, name)
}
