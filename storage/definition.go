package storage

import (
	"errors"

	"github.com/mylxsw/sync/meta"
	"github.com/siddontang/ledisdb/ledis"
)

var ErrNoSuchDefinition = errors.New("no such definition")

type DefinitionStore interface {
	Update(def meta.FileSyncGroup) error
	Get(id string) (meta.FileSyncGroup, error)
	All() ([]meta.FileSyncGroup, error)
}

type definitionStore struct {
	db *ledis.DB
}

func NewDefinitionStore(db *ledis.DB) DefinitionStore {
	return &definitionStore{db: db}
}

func (d *definitionStore) Update(def meta.FileSyncGroup) error {
	panic("implement me")
}

func (d *definitionStore) Get(id string) (meta.FileSyncGroup, error) {
	panic("implement me")
}

func (d *definitionStore) All() ([]meta.FileSyncGroup, error) {
	panic("implement me")
}
