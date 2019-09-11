package utils_test

import (
	"testing"

	"github.com/mylxsw/sync/utils"
	"github.com/stretchr/testify/assert"
)

func TestStringArrayUnique(t *testing.T) {
	arr := []string{
		"aaa",
		"bbb",
		"ccc",
		"aaa",
		"ddd",
		"ccc",
	}

	assert.EqualValues(t, 4, len(utils.StringArrayUnique(arr)))
}
