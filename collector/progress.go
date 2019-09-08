package collector

import (
	"sync"
)

type Progress struct {
	lock  sync.RWMutex
	total int
	max   int
}

func NewProgress(max int) *Progress {
	return &Progress{max: max}
}

func (p *Progress) SetTotal(total int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.total = total
}

func (p *Progress) Percentage() float32 {
	p.lock.RLock()
	defer p.lock.RUnlock()

	if p.total >= p.max {
		return 1
	}

	return float32(p.total) / float32(p.max)
}

func (p *Progress) Total() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.total
}

func (p *Progress) Max() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.max
}

func (p *Progress) Add(count int) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.total += count
}
