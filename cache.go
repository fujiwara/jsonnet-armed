package armed

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Cache struct {
	dir string
	ttl time.Duration
}

// NewCache creates a new cache instance
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		dir: getCacheDir(),
		ttl: ttl,
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

// Get retrieves a cached result if it exists and is not expired
func (c *Cache) Get(key string) (string, bool) {
	if c.ttl == 0 {
		return "", false
	}

	cachePath := filepath.Join(c.dir, key+".json")

	// Check file stats first
	stat, err := os.Stat(cachePath)
	if err != nil {
		return "", false
	}

	// Check if cache is expired
	if time.Since(stat.ModTime()) > c.ttl {
		// Cache is expired, remove it
		os.Remove(cachePath)
		return "", false
	}

	// Read the cached result directly
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return "", false
	}

	return string(data), true
}

// Set stores a result in the cache
func (c *Cache) Set(key string, result string) error {
	if c.ttl == 0 {
		return nil
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(c.dir, 0755); err != nil {
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

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		cachePath := filepath.Join(c.dir, entry.Name())
		stat, err := os.Stat(cachePath)
		if err != nil {
			continue
		}

		if time.Since(stat.ModTime()) > c.ttl {
			// Expired, remove it
			os.Remove(cachePath)
		}
	}

	return nil
}
