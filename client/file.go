package client

import (
	"github.com/mylxsw/sync/protocol"
)

type FileNeedSync struct {
	SaveFilePath string
	SyncOwner    bool
	SyncFile     bool
	Chmod        bool
	Type         protocol.Type
	RemoteFile   *protocol.File
}

func (fns FileNeedSync) NeedSync() bool {
	return fns.SyncFile || fns.SyncOwner || fns.Chmod
}

type FileNeedSyncs struct {
	files []FileNeedSync
}

func (fns *FileNeedSyncs) Add(n FileNeedSync) {
	fns.files = append(fns.files, n)
}

func (fns *FileNeedSyncs) All() []FileNeedSync {
	return fns.files
}
