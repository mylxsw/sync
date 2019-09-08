package collector

import (
	"encoding/json"
	"sync"

	"github.com/mylxsw/asteria/log"
)

// Collectors 用于存储当前活跃的所有collector
type Collectors struct {
	lock       sync.RWMutex
	collectors map[string]*Collector
}

func NewCollectors() *Collectors {
	return &Collectors{collectors: make(map[string]*Collector),}
}

func (cols *Collectors) Add(col *Collector) {
	cols.lock.Lock()
	defer cols.lock.Unlock()

	cols.collectors[col.jobID] = col
}

func (cols *Collectors) Remove(name string) {
	cols.lock.Lock()
	defer cols.lock.Unlock()

	delete(cols.collectors, name)
}

func (cols *Collectors) Get(id string) *Collector {
	cols.lock.RLock()
	defer cols.lock.RUnlock()

	return cols.collectors[id]
}

func (cols *Collectors) Names() []string {
	cols.lock.RLock()
	defer cols.lock.RUnlock()

	names := make([]string, 0)
	for key := range cols.collectors {
		names = append(names, key)
	}

	return names
}

// Collector 数据采集器，用于采集 Job 的输出
type Collector struct {
	lock       sync.RWMutex
	jobID      string
	collectors *Collectors
	Stages     []*Stage `json:"stages"`
}

// NewCollector 创建一个新的数据采集器
func NewCollector(collectors *Collectors, jobID string) *Collector {
	log.Debug("new collector created, collecting...")

	col := &Collector{Stages: make([]*Stage, 0), jobID: jobID, collectors: collectors}
	collectors.Add(col)

	return col
}

// Stage 创建一个 Stage
func (col *Collector) Stage(name string) *Stage {
	col.lock.Lock()
	defer col.lock.Unlock()

	log.Infof("---- stage %s ----", name)
	stage := &Stage{Name: name, Messages: make([]StageMessage, 0), col: col}
	col.Stages = append(col.Stages, stage)
	return stage
}

// HasError return whether there is an error message
func (col *Collector) HasError() bool {
	col.lock.RLock()
	defer col.lock.RUnlock()

	for _, s := range col.Stages {
		for _, m := range s.Messages {
			if m.Level == "ERROR" {
				return true
			}
		}
	}

	return false
}

// Build 转换为文本输出
func (col *Collector) Build() []byte {
	col.lock.Lock()
	defer col.lock.Unlock()

	for _, s := range col.Stages {
		s.Finish()
	}

	log.Debug("collect finished")
	res, _ := json.Marshal(col)
	return res
}

func (col *Collector) Finish() {
	col.lock.Lock()
	defer col.lock.Unlock()

	for _, s := range col.Stages {
		s.Finish()
	}

	col.collectors.Remove(col.jobID)
}

func (col *Collector) AllStages() []*Stage {
	col.lock.RLock()
	defer col.lock.RUnlock()

	return col.Stages
}
