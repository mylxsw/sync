package collector_test

import (
	"testing"

	"github.com/mylxsw/sync/collector"
	"github.com/stretchr/testify/assert"
)

func TestCollector_Build(t *testing.T) {
	collectors := collector.NewCollectors()

	coll := collector.NewCollector(collectors, "123")
	stage1 := coll.Stage("test1")
	stage1.Info("Hello, world")
	stage1.Info("Thanks")
	stage1.Error("Sorry, There are some errors")

	stage2 := coll.Stage("test2")
	stage2.Error("oops")

	assert.EqualValues(t, 3, len(coll.Stages[0].Messages))
	assert.Equal(t, 1, len(coll.Stages[1].Messages))

	assert.NotEmpty(t, coll.Build())

	assert.NotNil(t, collectors.Get("123"))
	coll.Finish()
	assert.Nil(t, collectors.Get("123"))
}
