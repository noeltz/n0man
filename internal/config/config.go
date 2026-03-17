package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the central n0man configuration `n0man.toml`
type Config struct {
	Version   int                          `toml:"version"`
	RemoteURL string                       `toml:"remote_url,omitempty"`
	LocalPath string                       `toml:"local_path,omitempty"`
	Settings  Settings                     `toml:"settings"`
	Security  SecurityConfig               `toml:"security"`
	Dotfiles  map[string]any               `toml:"dotfiles"`
	Overrides map[string]map[string]string `toml:"overrides,omitempty"`
}

type Settings struct {
	HousekeepingMaxBackups int  `toml:"housekeeping_max_backups"`
	RunBootstrapAfterInit  bool `toml:"run_bootstrap_after_init"`
}

type SecurityConfig struct {
	Enabled         bool   `toml:"enabled"`
	ScanContent     bool   `toml:"scan_content"`
	ExcludePatterns bool   `toml:"exclude_patterns"`
	Sensitivity     string `toml:"sensitivity"`
	FailOnSecrets   bool   `toml:"fail_on_secrets"`
	Interactive     bool   `toml:"interactive"`

	PatternConfig PatternConfig     `toml:"pattern_config,omitempty"`
	ContentScan   ContentScanConfig `toml:"content_scan,omitempty"`
	Allowlist     AllowlistConfig   `toml:"allowlist,omitempty"`
}

type PatternConfig struct {
	Custom []string `toml:"custom,omitempty"`
}

type ContentScanConfig struct {
	EntropyThreshold float64 `toml:"entropy_threshold"`
	MinSecretLength  int     `toml:"min_secret_length"`
	MaxFileSize      int     `toml:"max_file_size"`
	ScanBinaryFiles  bool    `toml:"scan_binary_files"`
	ContextWindow    int     `toml:"context_window"`
}

type AllowlistConfig struct {
	Patterns []string `toml:"patterns,omitempty"`
	Files    []string `toml:"files,omitempty"`
}

func (c *Config) GetPaths() map[string]string {
	paths := make(map[string]string)
	if c.Dotfiles == nil {
		c.Dotfiles = make(map[string]any)
	}
	for k, v := range c.Dotfiles {
		if k == "ignores" {
			continue
		}
		if s, ok := v.(string); ok {
			paths[k] = s
		}
	}
	return paths
}

func (c *Config) GetIgnores() map[string][]string {
	ignores := make(map[string][]string)
	if c.Dotfiles == nil {
		c.Dotfiles = make(map[string]any)
	}
	v, ok := c.Dotfiles["ignores"]
	if !ok {
		return ignores
	}

	m, ok := v.(map[string]any)
	if !ok {
		// Try map[string][]string just in case it was set manually
		if m2, ok := v.(map[string][]string); ok {
			return m2
		}
		return ignores
	}

	for k, iv := range m {
		switch parts := iv.(type) {
		case []string:
			ignores[k] = parts
		case []any:
			for _, p := range parts {
				if s, ok := p.(string); ok {
					ignores[k] = append(ignores[k], s)
				}
			}
		}
	}
	return ignores
}

func (c *Config) SetPath(name, path string) {
	if c.Dotfiles == nil {
		c.Dotfiles = make(map[string]any)
	}
	c.Dotfiles[name] = path
}

func (c *Config) SetIgnores(name string, patterns []string) {
	if c.Dotfiles == nil {
		c.Dotfiles = make(map[string]any)
	}

	var ignoresMap map[string]any
	if v, ok := c.Dotfiles["ignores"]; ok {
		ignoresMap, ok = v.(map[string]any)
		if !ok {
			// If it's map[string][]string, convert it
			if m2, ok := v.(map[string][]string); ok {
				ignoresMap = make(map[string]any)
				for k, val := range m2 {
					ignoresMap[k] = val
				}
			}
		}
	}

	if ignoresMap == nil {
		ignoresMap = make(map[string]any)
	}
	ignoresMap[name] = patterns
	c.Dotfiles["ignores"] = ignoresMap
}

func (c *Config) Delete(name string) {
	if c.Dotfiles == nil {
		return
	}
	delete(c.Dotfiles, name)
	if v, ok := c.Dotfiles["ignores"]; ok {
		if m, ok := v.(map[string]any); ok {
			delete(m, name)
			if len(m) == 0 {
				delete(c.Dotfiles, "ignores")
			} else {
				c.Dotfiles["ignores"] = m
			}
		}
	}
}

var ErrConfigNotFound = errors.New("n0man.toml not found")

// DefaultConfigPath returns the default path for the configuration file
func DefaultConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "n0man", "n0man.toml"), nil
}

// DefaultStorePath returns the default path for the internal store
func DefaultStorePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".local", "share", "n0man", "store"), nil
}

func DefaultConfig() Config {
	return Config{
		Version: 1,
		Settings: Settings{
			HousekeepingMaxBackups: 5,
			RunBootstrapAfterInit:  true,
		},
		Security: SecurityConfig{
			Enabled:         true,
			ScanContent:     true,
			ExcludePatterns: true,
			Sensitivity:     "medium",
			FailOnSecrets:   true,
			Interactive:     true,
			ContentScan: ContentScanConfig{
				EntropyThreshold: 4.5,
				MinSecretLength:  20,
				MaxFileSize:      10 * 1024 * 1024, // 10MB
				ScanBinaryFiles:  false,
				ContextWindow:    50,
			},
		},
		Dotfiles:  make(map[string]any),
		Overrides: make(map[string]map[string]string),
	}
}

// Load reads and parses the TOML config from the given path
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	// Initialize maps if nil
	if cfg.Dotfiles == nil {
		cfg.Dotfiles = make(map[string]any)
	}
	if cfg.Overrides == nil {
		cfg.Overrides = make(map[string]map[string]string)
	}

	return &cfg, nil
}

// GetTargetPath returns the target path for a dotfile, considering host-specific overrides
func (c *Config) GetTargetPath(name string) string {
	target, ok := c.GetPaths()[name]
	if !ok {
		return ""
	}

	hostname, _ := os.Hostname()
	if hostOverrides, ok := c.Overrides[hostname]; ok {
		if overridePath, ok := hostOverrides[name]; ok {
			return overridePath
		}
	}

	return target
}

// Save writes the config to the given path
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Create backup of existing config before overwriting
	if _, err := os.Stat(path); err == nil {
		// Config exists, create timestamped backup
		timestamp := time.Now().Format("20060102_150405")
		backupPath := fmt.Sprintf("%s.backup-%s", path, timestamp)

		existing, readErr := os.ReadFile(path)
		if readErr == nil {
			if writeErr := os.WriteFile(backupPath, existing, 0600); writeErr == nil {
				// Keep only last 5 backups
				cleanupOldBackups(path, 5)
			}
		}
	}

	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	// Write with secure permissions (user-only read/write)
	return os.WriteFile(path, b, 0600)
}

// cleanupOldBackups removes old backup files, keeping only the most recent ones
func cleanupOldBackups(configPath string, keep int) {
	if keep <= 0 {
		return
	}

	dir := filepath.Dir(configPath)
	baseName := filepath.Base(configPath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []string
	prefix := baseName + ".backup-"

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			backups = append(backups, entry.Name())
		}
	}

	// Sort by name (timestamp format ensures chronological order)
	sort.Strings(backups)

	// Remove oldest backups beyond the keep limit
	if len(backups) > keep {
		for i := 0; i < len(backups)-keep; i++ {
			_ = os.Remove(filepath.Join(dir, backups[i]))
		}
	}
}
