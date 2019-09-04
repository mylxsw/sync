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
		assert.EqualValues(t, "", files[0].Path)
	}
	{
		files, err := utils.AllFiles("./testcase/hello.link")
		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.EqualValues(t, "", files[0].Path)
	}
	{
		files, err := utils.AllFiles("./testcase/hello2")
		assert.NoError(t, err)
		assert.Len(t, files, 3)
		assert.EqualValues(t, "", files[0].Path)
	}
	{
		files, err := utils.AllFiles("./testcase")
		assert.NoError(t, err)
		assert.Len(t, files, 7)

		for _, f := range files {
			switch f.Path {
			case "":
				assert.Equal(t, utils.FileTypeDirectory, f.Type)
			case "hello":
				assert.Equal(t, utils.FileTypeDirectory, f.Type)
				assert.Empty(t, f.Symlink)
				assert.Empty(t, f.Checksum)
			case "hello/.hiddenfile":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "902fbdd2b1df0c4f70b4a5d23525e932", f.Checksum)
			case "hello/hello.md":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "446b211e6a3f7fdea10e34a6217599e8", f.Checksum)
			case "hello.link":
				assert.Equal(t, utils.FileTypeSymlink, f.Type)
				assert.Equal(t, "hello.txt", f.Symlink)
			case "hello.txt":
				assert.Equal(t, utils.FileTypeNormal, f.Type)
				assert.Equal(t, "63ef8d2059a401bfdbdefa864f6b9739", f.Checksum)
			case "hello2":
				assert.Equal(t, utils.FileTypeSymlink, f.Type)
				assert.Equal(t, "hello", f.Symlink)
			default:
				t.Errorf("file not matched: %s", f.Path)
			}
		}
	}
}
