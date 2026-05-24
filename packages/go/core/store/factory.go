package store

import (
	"fmt"

	"github.com/Nadil-K/dusk/packages/go/core"
)

func CreateStore(cfg core.DuskConfig) (*AsyncWriteBuffer, error) {
	var backend HitStore
	var err error

	switch cfg.Store.Backend {
	case "sqlite", "":
		path := cfg.Store.Path
		if path == "" {
			path = ".dusk/hits.db"
		}
		backend, err = NewSQLiteHitStore(path)
	case "redis":
		url := cfg.Store.URL
		if url == "" {
			url = "redis://localhost:6379"
		}
		prefix := cfg.Store.KeyPrefix
		if prefix == "" {
			prefix = "dusk"
		}
		ttlDays := cfg.Store.TTLDays
		if ttlDays == 0 {
			ttlDays = 90
		}
		backend, err = NewRedisHitStore(url, prefix, ttlDays)
	default:
		return nil, fmt.Errorf("unknown store backend: %s", cfg.Store.Backend)
	}

	if err != nil {
		return nil, err
	}

	buf := NewAsyncWriteBuffer(backend, 0, 0, 0)
	buf.Start()
	return buf, nil
}
