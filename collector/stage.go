package collector

import (
	"sync"
	"time"

	"github.com/mylxsw/asteria/log"
)

// Stage 采集阶段
type Stage struct {
	lock       sync.RWMutex
	Name       string         `json:"name"`
	Messages   []StageMessage `json:"messages"`
	Percentage float32        `json:"percentage"`
	Max        int            `json:"max"`
	Total      int            `json:"total"`
	progress   *Progress
	col        *Collector
}

type StageMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func (s *Stage) Progress(max int) *Progress {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.progress = NewProgress(max)
	s.Max = max
	return s.progress
}

func (s *Stage) Finish() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.progress != nil {
		s.Percentage = s.progress.Percentage()
		s.Total = s.progress.Total()

		s.progress = nil
	}
}

// Info 标准输出
func (s *Stage) Info(message string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Info(message)
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   message,
	})
}

// Error 错误输出
func (s *Stage) Error(message string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Error(message)
	s.Messages = append(s.Messages, StageMessage{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   message,
	})
}
