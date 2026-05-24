package core

type EndpointConfig struct {
	Path         string   `yaml:"path"`
	Methods      []string `yaml:"methods"`
	DeprecatedAt string   `yaml:"deprecated_at"`
	SunsetAt     *string  `yaml:"sunset_at"`
	Successor    *string  `yaml:"successor"`
	MigrationDoc *string  `yaml:"migration_doc"`
	Note         *string  `yaml:"note"`
}

type StoreConfig struct {
	Backend   string `yaml:"backend"`
	Path      string `yaml:"path"`
	URL       string `yaml:"url"`
	KeyPrefix string `yaml:"key_prefix"`
	TTLDays   int    `yaml:"ttl_days"`
}

type LogConfig struct {
	IdentifyBy []interface{} `yaml:"identify_by"`
}

type DuskConfig struct {
	Version   int              `yaml:"version"`
	Store     StoreConfig      `yaml:"store"`
	Log       LogConfig        `yaml:"log"`
	Endpoints []EndpointConfig `yaml:"endpoints"`
}
