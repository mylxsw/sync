package collector

import (
	"encoding/json"
	"time"

	"github.com/mylxsw/asteria/log"
)

// Collector 数据采集器，用于采集 Job 的输出
type Collector struct {
	Stages []*Stage `json:"stages"`
}

// NewCollector 创建一个新的数据采集器
func NewCollector() *Collector {
	log.Debug("new collector created, collecting...")
	return &Collector{Stages: make([]*Stage, 0)}
}

// Stage 采集阶段
type Stage struct {
	Name     string         `json:"name"`
	Messages []StageMessage `json:"messages"`
}

type StageMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Info 标准输出
func (s *Stage) Info(message string) {
	log.Info(message)
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   message,
	})
}

// Error 错误输出
func (s *Stage) Error(message string) {
	log.Error(message)
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   message,
	})
}

// Stage 创建一个 Stage
func (col *Collector) Stage(name string) *Stage {
	log.Infof("---- stage %s ----", name)
	stage := &Stage{Name: name, Messages: make([]StageMessage, 0)}
	col.Stages = append(col.Stages, stage)
	return stage
}

// Build 转换为文本输出
func (col *Collector) Build() []byte {
	log.Debug("collect finished")
	res, _ := json.Marshal(col)
	return res
}
