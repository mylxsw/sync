package collector_test

import (
	"testing"

	"github.com/mylxsw/sync/collector"
	"github.com/stretchr/testify/assert"
)

func TestCollector_Build(t *testing.T) {
	coll := collector.NewCollector()
	stage1 := coll.Stage("test1")
	stage1.Log("Hello, world")
	stage1.Log("Thanks")
	stage1.Error("Sorry, There are some errors")

	stage2 := coll.Stage("test2")
	stage2.Error("oops")

	assert.EqualValues(t, 3, len(coll.Stages[0].Messages))
	assert.Equal(t, 1, len(coll.Stages[1].Messages))

	assert.NotEmpty(t, coll.Build())

}
