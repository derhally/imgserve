package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "imgserve.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfig(t *testing.T) {
	path := writeConfig(t, `
port: "9090"
logLevel: DEBUG
cacheDir: cache
directories:
  - path: /mnt/media/photos
  - path: /mnt/backup/FamilyShare/adly/screensaver-images
    name: adly
`)
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" || cfg.LogLevel != "DEBUG" {
		t.Errorf("port/logLevel = %q/%q", cfg.Port, cfg.LogLevel)
	}
	if want := filepath.Join(filepath.Dir(path), "cache"); cfg.CacheDir != want {
		t.Errorf("cacheDir = %q, want %q (relative to config file)", cfg.CacheDir, want)
	}

	sources, err := cfg.ImageSources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []ImageSource{
		{Name: "photos", Dir: "/mnt/media/photos"},
		{Name: "adly", Dir: "/mnt/backup/FamilyShare/adly/screensaver-images"},
	}
	if len(sources) != len(want) {
		t.Fatalf("got %d sources, want %d", len(sources), len(want))
	}
	for i := range want {
		if sources[i] != want[i] {
			t.Errorf("source %d = %+v, want %+v", i, sources[i], want[i])
		}
	}
}

func TestLoadConfigDefault(t *testing.T) {
	cfg, err := LoadConfig(writeConfig(t, `
directories:
  - path: /mnt/media/photos
  - path: /mnt/ss
    name: adly
    default: true
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sources, err := cfg.ImageSources()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sources[0].Default || !sources[1].Default {
		t.Errorf("default flags = %v, %v; want false, true", sources[0].Default, sources[1].Default)
	}
	if got := DefaultSourceIndex(sources); got != 1 {
		t.Errorf("DefaultSourceIndex = %d, want 1", got)
	}
}

func TestLoadConfigNumericPort(t *testing.T) {
	cfg, err := LoadConfig(writeConfig(t, "port: 8080\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("port = %q, want 8080", cfg.Port)
	}
}

func TestLoadConfigEmptyFile(t *testing.T) {
	cfg, err := LoadConfig(writeConfig(t, ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Directories) != 0 {
		t.Errorf("expected no directories, got %d", len(cfg.Directories))
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := map[string]string{
		"unknown key":   "prot: 8080\n",
		"missing path":  "directories:\n  - name: adly\n",
		"bad yaml":      "directories: [\n",
		"wrong type":    "directories: photos\n",
		"unknown field": "directories:\n  - path: /x\n    nmae: y\n",
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadConfig(writeConfig(t, content)); err == nil {
				t.Error("expected error")
			}
		})
	}

	if _, err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestConfigImageSourcesInvalidName(t *testing.T) {
	cfg := Config{Directories: []DirectoryConfig{{Path: "/mnt/x", Name: "bad name"}}}
	if _, err := cfg.ImageSources(); err == nil {
		t.Error("expected error for invalid name")
	}
}
