package storage

import (
	"errors"

	"github.com/siddontang/ledisdb/ledis"
)

var ErrNoSuchSetting = errors.New("no such setting")

const (
	GlobalNamespace    = "global"
	SyncActionSetting  = "sync-action"
)

// SettingFactory is a factory for creating settingStore
type SettingFactory interface {
	Namespace(namespace string) SettingStore
}

type settingFactory struct {
	db *ledis.DB
}

func NewSettingFactory(db *ledis.DB) *settingFactory {
	return &settingFactory{db: db}
}

func (s settingFactory) Namespace(namespace string) SettingStore {
	return &settingStore{db: s.db, namespace: []byte(namespace)}
}

// SettingStore is a store for settings
type SettingStore interface {
	Update(key string, payload []byte) error
	Get(key string) ([]byte, error)
}

type settingStore struct {
	db        *ledis.DB
	namespace []byte
}

func (s *settingStore) Update(key string, payload []byte) error {
	_, err := s.db.HSet(s.namespace, []byte(key), payload)
	return err
}

func (s *settingStore) Get(key string) ([]byte, error) {
	rs, err := s.db.HGet(s.namespace, []byte(key))
	if err != nil {
		return rs, err
	}

	if rs == nil {
		return rs, ErrNoSuchSetting
	}

	return rs, err
}
