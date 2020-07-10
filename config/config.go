package config

import (
	"github.com/kelseyhightower/envconfig"
)

// Config is the filing resource handler config
type Config struct {
	BindAddr                  string `envconfig:"BIND_ADDR"                  json:"-"`
	DatasetIndex              string `envconfig:"DATASET_SEARCH_INDEX"`
	ElasticSearchAPIURL       string `envconfig:"ELASTIC_SEARCH_URL"         json:"-"`
	MaxSearchResultsOffset    int    `envconfig:"MAX_SEARCH_RESULTS_OFFSET"`
	SignElasticsearchRequests bool   `envconfig:"SIGN_ELASTICSEARCH_REQUESTS"`
	TaxonomyFilename          string `envconfig:"TAXONOMY_FILENAME"`
}

var cfg *Config

// Get configures the application and returns the configuration
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                  ":10200",
		DatasetIndex:              "dataset-test",
		ElasticSearchAPIURL:       "http://localhost:9200",
		MaxSearchResultsOffset:    1000,
		SignElasticsearchRequests: false,
		TaxonomyFilename:          "taxonomy/taxonomy.json",
	}

	return cfg, envconfig.Process("", cfg)
}
