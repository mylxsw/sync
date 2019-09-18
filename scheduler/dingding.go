package scheduler

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/go-toolkit/period_job"
	"github.com/mylxsw/sync/storage"
	"github.com/mylxsw/sync/utils/ding"
)

// DingdingConsumer is a job used to consuming dingding jobs
type DingdingConsumer struct {
	qs          storage.QueueStore
	maxSendSize int
}

// NewDingdingConsumer create a new DingdingConsumer
func NewDingdingConsumer(cc *container.Container) period_job.Job {
	consumer := DingdingConsumer{maxSendSize: 5,}
	cc.MustResolve(func(qsf storage.QueueStoreFactory) {
		consumer.qs = qsf.Queue(storage.QueueDingding)
	})

	return &consumer
}

func (d *DingdingConsumer) Handle() {
	sendSize := 0
	for sendSize < d.maxSendSize {
		sendSize++

		payload, err := d.qs.Dequeue(0)
		if err != nil {
			log.Errorf("dequeue dingding jobs failed: %s", err)
			return
		}

		// if payload is nil, the queue is empty
		if payload == nil {
			return
		}

		var dMessage ding.DingdingMessage
		if err := dMessage.Decode(payload); err != nil {
			log.Errorf("decode dingding message failed: %s", err)
			return
		}

		client := ding.NewDingding(dMessage.Token)
		if err := client.Send(dMessage.Message); err != nil {
			log.WithFields(log.Fields{
				"ding": dMessage,
			}).Errorf("send dingding message failed: %s", err)
			return
		}

		log.WithFields(log.Fields{
			"ding": dMessage,
		}).Debug("send dingding message ok")
	}
}
