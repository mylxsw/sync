package collector

import (
	"fmt"
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
	Index     int       `json:"index"`
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

func (s *Stage) GetProgress() *Progress {
	s.lock.RLock()
	defer s.lock.RUnlock()

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

	// log.Info(message)
	msg := StageMessage{
		Index:     s.col.Index(),
		Timestamp: time.Now(),
		Level:     "INFO",
		Message:   message,
	}
	s.Messages = append(s.Messages, msg)
}

func (s *Stage) Infof(format string, a ...interface{}) {
	s.Info(fmt.Sprintf(format, a...))
}

func (s *Stage) Errorf(format string, a ...interface{}) {
	s.Error(fmt.Sprintf(format, a...))
}

// Error 错误输出
func (s *Stage) Error(message string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Error(message)
	msg := StageMessage{
		Index:     s.col.Index(),
		Timestamp: time.Now(),
		Level:     "ERROR",
		Message:   message,
	}
	s.Messages = append(s.Messages, msg)
}
