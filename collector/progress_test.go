package collector_test

import (
	"testing"

	"github.com/mylxsw/sync/collector"
	"github.com/stretchr/testify/assert"
)

func TestNewProgress(t *testing.T) {
	p := collector.NewProgress(1042)

	p.SetTotal(100)
	assert.True(t, p.Percentage() < 0.1 && p.Percentage() > 0.09)

	p.Add(100)
	assert.True(t, p.Percentage() < 0.2 && p.Percentage() > 0.19)

	assert.EqualValues(t, 200, p.Total())
}
