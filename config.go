package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Views []string
}

func ReadConfig(path string) (*Config, error) {
	var config Config
	f, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(f, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
