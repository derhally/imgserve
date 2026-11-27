package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config is the content of the YAML file passed with -config.
type Config struct {
	Port        string            `yaml:"port"`
	LogLevel    string            `yaml:"logLevel"`
	CacheDir    string            `yaml:"cacheDir"`
	CertFile    string            `yaml:"certFile"`
	CertKeyFile string            `yaml:"certKeyFile"`
	Directories []DirectoryConfig `yaml:"directories"`
}

// DirectoryConfig is one served image directory.
type DirectoryConfig struct {
	Path string `yaml:"path"` // directory holding the images
	Name string `yaml:"name"` // URL path name; defaults to the directory's base name
}

// LoadConfig reads a YAML config file. Unknown keys are rejected so typos
// don't go unnoticed. Relative paths are resolved against the directory
// that holds the config file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Config{}, err
	}
	baseDir := filepath.Dir(absPath)

	resolve := func(p string) string {
		if p == "" || filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(baseDir, p)
	}

	cfg.CacheDir = resolve(cfg.CacheDir)
	cfg.CertFile = resolve(cfg.CertFile)
	cfg.CertKeyFile = resolve(cfg.CertKeyFile)
	for i, d := range cfg.Directories {
		if d.Path == "" {
			return Config{}, fmt.Errorf("config file %s: directory %d has no path", path, i+1)
		}
		cfg.Directories[i].Path = resolve(d.Path)
	}

	return cfg, nil
}

// ImageSources converts the configured directories into image sources.
func (c Config) ImageSources() ([]ImageSource, error) {
	sources := make([]ImageSource, 0, len(c.Directories))
	for _, d := range c.Directories {
		source, err := NewImageSource(d.Name, d.Path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}
