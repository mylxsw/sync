package storage

import (
	"errors"

	"github.com/siddontang/ledisdb/ledis"
)

var ErrNoSuchJob = errors.New("no such job")

type FailedJobStore interface {
	Add(id string, data []byte) error
	All() ([][]byte, error)
	Get(id string) ([]byte, error)
	Delete(id string) error
}

type failedJobStore struct {
	db   *ledis.DB
	name []byte
}

func NewFailedJobStore(db *ledis.DB) FailedJobStore {
	return &failedJobStore{db: db, name: []byte("file-sync-failed")}
}

func (f *failedJobStore) Add(id string, data []byte) error {
	_, err := f.db.HSet(f.name, []byte(id), data)
	return err
}

func (f *failedJobStore) All() ([][]byte, error) {
	pairs, err := f.db.HGetAll(f.name)
	if err != nil {
		return nil, err
	}

	res := make([][]byte, 0)
	for _, p := range pairs {
		res = append(res, p.Value)
	}

	return res, nil
}

func (f *failedJobStore) Get(id string) ([]byte, error) {
	rs, err := f.db.HGet(f.name, []byte(id))
	if err != nil {
		return nil, err
	}

	if rs == nil {
		return nil, ErrNoSuchJob
	}

	return rs, nil
}

func (f *failedJobStore) Delete(id string) error {
	_, err := f.db.HDel(f.name, []byte(id))
	return err
}
