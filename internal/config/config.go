package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ScriptType string

const (
	TypeOneshot ScriptType = "oneshot"
	TypeService ScriptType = "service"
)

type ArgDef struct {
	Name     string `json:"name"`
	Flag     string `json:"flag"`
	Default  string `json:"default"`
	Required bool   `json:"required"`
}

type ScriptMeta struct {
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Desc      string     `json:"desc"`
	Type      ScriptType `json:"type"`
	Language  string     `json:"language"`
	Commands  []string   `json:"commands"`
	PidFile   string     `json:"pidFile"`
	Args      []ArgDef   `json:"args"`
	Tags      []string   `json:"tags"`
	LastRunAt int64      `json:"lastRunAt"`
	UpdatedAt int64      `json:"updatedAt"`
}

type Directory struct {
	Path        string `json:"path"`
	IsScriptDir bool   `json:"isScriptDir"`
	ExportName  string `json:"exportName"`
}

type ExportVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Config struct {
	Directories  []Directory           `json:"directories"`
	Exports      []ExportVar           `json:"exports"`
	OverrideMeta map[string]ScriptMeta `json:"overrideMeta"`
}

func defaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Hatch", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{OverrideMeta: map[string]ScriptMeta{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.OverrideMeta == nil {
		cfg.OverrideMeta = map[string]ScriptMeta{}
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := defaultPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
