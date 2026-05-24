package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type rawDusk struct {
	Dusk struct {
		Version int         `yaml:"version"`
		Store   StoreConfig `yaml:"store"`
		Log     struct {
			IdentifyBy []interface{} `yaml:"identify_by"`
		} `yaml:"log"`
		Endpoints []struct {
			Path         string   `yaml:"path"`
			Methods      []string `yaml:"methods"`
			DeprecatedAt string   `yaml:"deprecated_at"`
			SunsetAt     *string  `yaml:"sunset_at"`
			Successor    *string  `yaml:"successor"`
			MigrationDoc *string  `yaml:"migration_doc"`
			Note         *string  `yaml:"note"`
		} `yaml:"endpoints"`
	} `yaml:"dusk"`
}

func LoadConfig(path string) (DuskConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DuskConfig{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var raw rawDusk
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return DuskConfig{}, fmt.Errorf("parse config: %w", err)
	}

	d := raw.Dusk

	store := d.Store
	if store.Backend == "" {
		store.Backend = "sqlite"
	}
	if store.Path == "" {
		store.Path = ".dusk/hits.db"
	}
	// Resolve SQLite path relative to the config file so the DB lands next to
	// dusk.yaml regardless of the process working directory.
	if store.Backend == "sqlite" && !filepath.IsAbs(store.Path) {
		store.Path = filepath.Join(filepath.Dir(path), store.Path)
	}
	if store.KeyPrefix == "" {
		store.KeyPrefix = "dusk"
	}
	if store.TTLDays == 0 {
		store.TTLDays = 90
	}

	endpoints := make([]EndpointConfig, 0, len(d.Endpoints))
	for _, ep := range d.Endpoints {
		methods := ep.Methods
		if len(methods) == 0 {
			methods = []string{"GET"}
		}
		upper := make([]string, len(methods))
		for i, m := range methods {
			upper[i] = strings.ToUpper(m)
		}
		endpoints = append(endpoints, EndpointConfig{
			Path:         ep.Path,
			Methods:      upper,
			DeprecatedAt: ep.DeprecatedAt,
			SunsetAt:     ep.SunsetAt,
			Successor:    ep.Successor,
			MigrationDoc: ep.MigrationDoc,
			Note:         ep.Note,
		})
	}

	return DuskConfig{
		Version:   d.Version,
		Store:     store,
		Log:       LogConfig{IdentifyBy: d.Log.IdentifyBy},
		Endpoints: endpoints,
	}, nil
}
