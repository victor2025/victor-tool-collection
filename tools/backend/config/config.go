package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DBType     string `json:"db_type"`
	DSN        string `json:"dsn"`
	ServerPort string `json:"server_port"`
	ScoresDir  string `json:"scores_dir"`
}

func Load() (*Config, error) {
	// Find config.json relative to executable
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}
	configPath := filepath.Join(filepath.Dir(exe), "config.json")

	// Also check project root
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config.json: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config.json: %w", err)
	}

	// Validate required fields
	if cfg.DSN == "" {
		return nil, fmt.Errorf("config.json: dsn is required")
	}
	if cfg.DBType == "" {
		cfg.DBType = "postgres"
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8003"
	}
	if cfg.ScoresDir == "" {
		// Default: <repo>/data/scores, resolved relative to the executable
		// (exe lives in tools/backend/, so go up two levels).
		cfg.ScoresDir = filepath.Join(filepath.Dir(exe), "..", "..", "data", "scores")
	}
	if abs, err := filepath.Abs(cfg.ScoresDir); err == nil {
		cfg.ScoresDir = abs
	}

	return &cfg, nil
}
