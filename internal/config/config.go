package config

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rakunlabs/chu"
	"github.com/rakunlabs/chu/loader"
	"github.com/rakunlabs/chu/loader/loaderenv"
	"github.com/rakunlabs/chu/loader/loaderfile"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Binary settings
	BinPath      string `yaml:"bin_path"`     // Default: ~/bin
	Architecture string `yaml:"architecture"` // auto, amd64, arm64

	// Command behavior
	ShowExplanation bool `yaml:"show_explanation"` // Show command explanation before execution

	// UI settings
	Theme   string        `yaml:"theme"`   // default, dark, light
	Timeout time.Duration `yaml:"timeout"` // Download timeout in seconds

	// Network
	ProxyURL string `yaml:"proxy_url"` // HTTP proxy for downloads
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		BinPath:         filepath.Join(homeDir, "bin"),
		Architecture:    "auto",
		ShowExplanation: true,
		Theme:           "default",
		Timeout:         30 * time.Second,
		ProxyURL:        "",
	}
}

// GetArchitecture returns the effective architecture (resolves "auto")
func (c *Config) GetArchitecture() string {
	if c.Architecture == "auto" {
		arch := runtime.GOARCH
		// Normalize common architectures
		switch arch {
		case "amd64", "x86_64":
			return "amd64"
		case "arm64", "aarch64":
			return "arm64"
		default:
			return arch
		}
	}
	return c.Architecture
}

// GetOS returns the current operating system
func (c *Config) GetOS() string {
	os := runtime.GOOS
	switch os {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return os
	}
}

func Load(ctx context.Context) (*Config, error) {
	cfg := DefaultConfig()
	if err := chu.Load(
		ctx, "kup", cfg,
		chu.WithTag("yaml"),
		chu.WithDisableLoader(loader.NameHTTP),
		chu.WithLoaderOption(loaderenv.New(
			loaderenv.WithPrefix("KUP_"),
			loaderenv.WithEnvFile(),
			loaderenv.WithCheckConfigEnvFile(false),
		)),
		chu.WithLoaderOption(loaderfile.New(
			loaderfile.WithCheckCurrentFolder(false),
			loaderfile.WithCheckEnv(false),
			loaderfile.WithFolders(getConfigDir()),
		)),
	); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the configuration to the config file
func (c *Config) Save() error {
	configDir := getConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "kup.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func getConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "kup")
}
