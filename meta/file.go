package meta

import (
	"github.com/mylxsw/sync/protocol"
)

type FileNeedSync struct {
	SaveFilePath string
	SyncOwner    bool
	SyncFile     bool
	Chmod        bool
	Delete       bool
	Type         protocol.Type
	RemoteFile   *protocol.File
}

func (fns FileNeedSync) NeedSync() bool {
	return fns.SyncFile || fns.SyncOwner || fns.Chmod || fns.Delete
}

type FileNeedSyncs struct {
	Files []FileNeedSync
}

func (fns *FileNeedSyncs) Add(n FileNeedSync) {
	fns.Files = append(fns.Files, n)
}

func (fns *FileNeedSyncs) All() []FileNeedSync {
	return fns.Files
}
