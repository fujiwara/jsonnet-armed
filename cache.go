package armed

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cache struct {
	dir      string
	ttl      time.Duration
	staleTTL time.Duration
}

// NewCache creates a new cache instance
func NewCache(ttl time.Duration, staleTTL time.Duration) *Cache {
	return &Cache{
		dir:      getCacheDir(),
		ttl:      ttl,
		staleTTL: staleTTL,
	}
}

// getCacheDir returns the cache directory path following XDG Base Directory specification
func getCacheDir() string {
	// Check XDG_CACHE_HOME first
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return filepath.Join(cacheHome, "jsonnet-armed")
	}

	// Fall back to $HOME/.cache
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".cache", "jsonnet-armed")
	}

	// Last resort: use temporary directory
	return filepath.Join(os.TempDir(), "jsonnet-armed-cache")
}

// GenerateCacheKey creates a unique cache key based on input parameters and content
func (c *Cache) GenerateCacheKey(cli *CLI, content []byte) (string, error) {
	hasher := sha256.New()

	// Marshal the CLI configuration to JSON for consistent hashing
	// Private fields (writer, cacheKey) are automatically ignored
	cliJSON, err := json.Marshal(cli)
	if err != nil {
		return "", fmt.Errorf("failed to marshal CLI for cache key: %w", err)
	}

	// Hash the CLI configuration
	hasher.Write(cliJSON)

	// Add absolute path separately for files (not stdin) to ensure uniqueness
	if cli.Filename != "-" {
		absPath, err := filepath.Abs(cli.Filename)
		if err != nil {
			return "", err
		}
		hasher.Write([]byte(absPath))
	}

	// Hash the content
	hasher.Write(content)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// Get retrieves a cached result if it exists and is not expired (deprecated)
// Use GetWithStale instead for stale cache support
func (c *Cache) Get(key string) (string, bool) {
	content, isStale, exists := c.GetWithStale(key)
	if exists && !isStale {
		return content, true
	}
	return "", false
}

// GetWithStale retrieves a cached result with stale status information
// Returns (content, isStale, exists)
func (c *Cache) GetWithStale(key string) (string, bool, bool) {
	if c.ttl == 0 {
		return "", false, false
	}

	cachePath := filepath.Join(c.dir, key+".json")

	// Check file stats first
	stat, err := os.Stat(cachePath)
	if err != nil {
		return "", false, false
	}

	age := time.Since(stat.ModTime())

	// Check fresh cache first (within TTL)
	if age <= c.ttl {
		data, err := os.ReadFile(cachePath)
		if err != nil {
			slog.Warn("Failed to read fresh cache file",
				"error", err.Error(),
				"cache_path", cachePath,
				"cache_key", key[:8]+"...")
			return "", false, false
		}
		return string(data), false, true
	}

	// Check stale cache (within staleTTL)
	if c.staleTTL > 0 && age <= c.staleTTL {
		data, err := os.ReadFile(cachePath)
		if err != nil {
			slog.Warn("Failed to read stale cache file",
				"error", err.Error(),
				"cache_path", cachePath,
				"cache_key", key[:8]+"...")
			return "", false, false
		}
		return string(data), true, true
	}

	// Completely expired - remove it
	if err := os.Remove(cachePath); err != nil {
		slog.Warn("Failed to remove expired cache file",
			"error", err.Error(),
			"cache_path", cachePath,
			"cache_key", key[:8]+"...")
	}
	return "", false, false
}

// Set stores a result in the cache
func (c *Cache) Set(key string, result string) error {
	if c.ttl == 0 {
		return nil
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		slog.Warn("Failed to create cache directory",
			"error", err.Error(),
			"cache_dir", c.dir)
		return err
	}

	// Write the result directly to cache file
	// Use 0600 permissions to protect potentially sensitive information
	cachePath := filepath.Join(c.dir, key+".json")
	return writeFileAtomic(cachePath, []byte(result), 0600)
}

// Clean removes expired cache entries
func (c *Cache) Clean() error {
	if c.ttl == 0 {
		return nil
	}

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Determine maximum age to keep cache files
	// Use staleTTL if it's longer than ttl, otherwise use ttl
	maxAge := c.ttl
	if c.staleTTL > 0 && c.staleTTL > c.ttl {
		maxAge = c.staleTTL
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		cachePath := filepath.Join(c.dir, entry.Name())
		stat, err := os.Stat(cachePath)
		if err != nil {
			continue
		}

		if time.Since(stat.ModTime()) > maxAge {
			// Completely expired, remove it
			if err := os.Remove(cachePath); err != nil {
				slog.Warn("Failed to remove expired cache file during cleanup",
					"error", err.Error(),
					"cache_path", cachePath)
			}
		}
	}

	return nil
}
