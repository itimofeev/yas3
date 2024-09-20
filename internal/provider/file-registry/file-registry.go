package file_registry

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-playground/validator/v10"
)

type Config struct {
	DBPath string `validate:"required"`
}

// Registry contains information about uploaded files. For example on which server which file part is stored.
type Registry struct {
	db *badger.DB
}

func New(cfg Config) (*Registry, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	db, err := badger.Open(badger.DefaultOptions(cfg.DBPath))
	if err != nil {
		return nil, err
	}

	return &Registry{
		db: db,
	}, nil
}

func (r *Registry) SaveFileParts(fileID string, serverIDs []string) error {
	return r.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(fileID), []byte(strings.Join(serverIDs, ",")))
		return err
	})
}

func (r *Registry) GetFileParts(fileID string) ([]string, error) {
	var serverIDs []string
	err := r.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(fileID))
		if err != nil {
			return err
		}

		valueCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		serverIDs = strings.Split(string(valueCopy), ",")
		return nil
	})

	return serverIDs, err
}

func (r *Registry) IsFileExists(fileID string) bool {
	err := r.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(fileID))
		return err
	})

	return !errors.Is(err, badger.ErrKeyNotFound)
}

func (r *Registry) Close() error {
	return r.db.Close()
}
