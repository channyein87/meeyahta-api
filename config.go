package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type config struct {
	APIKey string `json:"apikey"`
}

func loadConfig(path string) (config, error) {
	f, err := os.Open(path)
	if err != nil {
		return config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return config{}, fmt.Errorf("decode config: %w", err)
	}

	if cfg.APIKey == "" {
		return config{}, fmt.Errorf("apikey is required")
	}

	return cfg, nil
}
