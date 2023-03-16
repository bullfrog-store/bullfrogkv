package storage

import (
	"errors"

	"github.com/cockroachdb/pebble"
)

type Engine interface {
	WriteData(m Modify) error
	WriteMeta(m Modify) error
	ReadData(key []byte) ([]byte, error)
	ReadMeta(key []byte) ([]byte, error)
	DataSnapshot() ([]byte, error)
	Close() error
	Destroy() error
}

const (
	DataAbsDir = "/data"
	MetaAbsDir = "/meta"
)

var (
	ErrNotFound      = errors.New("key does not exist")
	ErrUnknownModify = errors.New("unknown modify type")
)

type Config struct {
	Path     string          // must be filled
	DataOpts *pebble.Options // optional
	MetaOpts *pebble.Options // optional
}

func NewDefaultConfig(path string) *Config {
	return &Config{Path: path}
}

type PapiEngine struct {
	// data stores the data of clients.
	data *papi
	// meta stores the metadata of the system.
	meta *papi

	path string
}

func New(config *Config) (Engine, error) {
	var err error

	e := &PapiEngine{path: config.Path}

	e.data, err = newPapi(e.path+DataAbsDir, config.DataOpts)
	if err != nil {
		return nil, err
	}
	e.meta, err = newPapi(e.path+MetaAbsDir, config.MetaOpts)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (e *PapiEngine) WriteData(m Modify) error {
	switch m.Data.(type) {
	case Put:
		return e.data.set(m.Key(), m.Value(), m.Sync())
	case Delete:
		return e.data.delete(m.Key(), m.Sync())
	default:
		return ErrUnknownModify
	}
}

func (e *PapiEngine) WriteMeta(m Modify) error {
	switch m.Data.(type) {
	case Put:
		return e.meta.set(m.Key(), m.Value(), m.Sync())
	case Delete:
		return e.meta.delete(m.Key(), m.Sync())
	default:
		return ErrUnknownModify
	}
}

func (e *PapiEngine) ReadData(key []byte) ([]byte, error) {
	return e.data.get(key)
}

func (e *PapiEngine) ReadMeta(key []byte) ([]byte, error) {
	return e.meta.get(key)
}

func (e *PapiEngine) DataSnapshot() ([]byte, error) {
	return e.data.snapshot()
}

func (e *PapiEngine) Close() error {
	if err := e.data.close(); err != nil {
		return err
	}
	if err := e.meta.close(); err != nil {
		return err
	}
	return nil
}

func (e *PapiEngine) Destroy() error {
	if err := e.Close(); err != nil {
		return err
	}
	if err := removeAll(e.path); err != nil {
		return err
	}
	return nil
}
