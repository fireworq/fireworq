package config

import (
	"os"
	"strings"
	"sync"
)

type config struct {
	sync.RWMutex
	c map[string]string
}

var cached = config{
	c: make(map[string]string),
}

// Get returns the current configuration value of a key.
//
// If it has no specific value, it falls back to a value of
// environment variable starting with "FIREWORQ_" and then a default
// value which is returned from GetDefault.
func Get(key string) string {
	cached.RLock()
	v, ok := cached.c[key]
	cached.RUnlock()
	if ok {
		return v
	}

	envKey := "FIREWORQ_" + strings.ToUpper(key)
	v = os.Getenv(envKey)
	if v == "" {
		v = GetDefault(key)
	}
	cached.Lock()
	cached.c[key] = v
	cached.Unlock()
	return v
}

// GetDefault returns the default configuration value of a key.
func GetDefault(key string) string {
	cached.RLock()
	defer cached.RUnlock()
	item, ok := defaultConf[key]
	if ok {
		return item.defaultValue
	}
	return ""
}

// Set sets the current configuration value of a key.
func Set(k, v string) {
	cached.Lock()
	cached.c[k] = v
	cached.Unlock()
}

// SetDefault sets the default configuration value of a key.
func SetDefault(k, v string) {
	cached.Lock()
	defer cached.Unlock()
	item, ok := defaultConf[k]
	if ok {
		item.defaultValue = v
	} else {
		defaultConf[k] = &configItem{defaultValue: v}
	}
}

// Locally overrides the current configuration value of a key in a block.
//
// This is not goroutine safe and should only be used in tests.
func Locally(k, v string, block func()) {
	original := Get(k)
	Set(k, v)
	defer func() { Set(k, original) }()
	block()
}

// Keys returns a list of configuration keys.
func Keys() []string {
	cached.RLock()
	defer cached.RUnlock()

	keys := make([]string, 0, len(cached.c))
	for k := range defaultConf {
		keys = append(keys, k)
	}
	return keys
}
