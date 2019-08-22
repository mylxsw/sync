package utils_test

import (
	"testing"

	"github.com/mylxsw/sync/utils"
	"github.com/stretchr/testify/assert"
)

func TestAllFiles(t *testing.T) {
	{
		files, err := utils.AllFiles("./testcase/hello.txt")
		assert.NoError(t, err)
		assert.Len(t, files, 1)
	}
	{
		files, err := utils.AllFiles("./testcase")
		assert.NoError(t, err)
		assert.Len(t, files, 7)

		for _, f := range files {
			switch f.Path {
			case "./testcase":
				assert.Equal(t, utils.FileTypeDirectory, f.Type)
			case "testcase/hello":
				assert.Equal(t, utils.FileTypeDirectory, f.Type)
				assert.Empty(t, f.Symlink)
				assert.Empty(t, f.Checksum)
			case "testcase/hello/.hiddenfile":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "902fbdd2b1df0c4f70b4a5d23525e932", f.Checksum)
			case "testcase/hello/hello.md":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "446b211e6a3f7fdea10e34a6217599e8", f.Checksum)
			case "testcase/hello.link":
				assert.Equal(t, utils.FileTypeSymlink, f.Type)
				assert.Equal(t, "hello.txt", f.Symlink)
			case "testcase/hello.txt":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "63ef8d2059a401bfdbdefa864f6b9739", f.Checksum)
			case "testcase/hello2":
				assert.Equal(t, utils.FileTypeSymlink, f.Type)
				assert.Equal(t, "hello", f.Symlink)
			default:
				t.Errorf("file not matched: %s", f.Path)
			}
		}
	}
}
