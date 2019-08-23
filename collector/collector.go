package collector

import (
	"encoding/json"
	"time"
)

// Collector 数据采集器，用于采集 Job 的输出
type Collector struct {
	Stages []*Stage `json:"stages"`
}

// NewCollector 创建一个新的数据采集器
func NewCollector() *Collector {
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

// Log 标准输出
func (s *Stage) Log(message string) {
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "LOG",
		Message:   message,
	})
}

// Error 错误输出
func (s *Stage) Error(message string) {
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "ERR",
		Message:   message,
	})
}

// Stage 创建一个 Stage
func (col *Collector) Stage(name string) *Stage {
	stage := &Stage{Name: name, Messages: make([]StageMessage, 0)}
	col.Stages = append(col.Stages, stage)
	return stage
}

// Build 转换为文本输出
func (col *Collector) Build() []byte {
	res, _ := json.Marshal(col)
	return res
}
