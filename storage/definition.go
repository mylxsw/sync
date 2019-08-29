package storage

import (
	"errors"

	"github.com/mylxsw/sync/meta"
	"github.com/siddontang/ledisdb/ledis"
)

// ErrNoSuchDefinition 没有找到指定的定义
var ErrNoSuchDefinition = errors.New("no such definition")

// DefinitionStore 同步定义存储接口
type DefinitionStore interface {
	Update(def meta.FileSyncGroup) error
	Get(name string) (*meta.FileSyncGroup, error)
	All() ([]meta.FileSyncGroup, error)
}

type definitionStore struct {
	db *ledis.DB
}

func NewDefinitionStore(db *ledis.DB) DefinitionStore {
	return &definitionStore{db: db}
}

func (d *definitionStore) Update(def meta.FileSyncGroup) error {
	_, err := d.db.HSet([]byte("definition"), []byte(def.Name), def.Encode())
	return err
}

func (d *definitionStore) Get(name string) (*meta.FileSyncGroup, error) {
	data, err := d.db.HGet([]byte("definition"), []byte(name))
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, ErrNoSuchDefinition
	}

	var rs meta.FileSyncGroup
	return &rs, rs.Decode(data)
}

func (d *definitionStore) All() ([]meta.FileSyncGroup, error) {
	pairs, err := d.db.HGetAll([]byte("definition"))
	if err != nil {
		return nil, err
	}

	groups := make([]meta.FileSyncGroup, 0)
	for _, p := range pairs {
		var rs meta.FileSyncGroup
		if err := rs.Decode(p.Value); err != nil {
			return nil, err
		}

		groups = append(groups, rs)
	}

	return groups, nil
}
