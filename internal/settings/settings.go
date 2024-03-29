package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	Konfig        *koanf.Koanf
	settingsFile  string
	CacheDuration int64 = 30 * 60

	ErrNotFound = errors.New("not found")
	ErrExpired  = errors.New("cache entry expired")
)

func Config() {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("error getting config folder: %v\n", err)
		os.Exit(1)
	}

	path := filepath.Join(dir, "platoon")
	os.MkdirAll(path, 0755)
	settingsFile = filepath.Join(path, "settings.json")

	Konfig = koanf.New(".")
	if err := Konfig.Load(file.Provider(settingsFile), koanfJson.Parser()); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("error in config file (%s): %v\n", settingsFile, err)
		os.Exit(1)
	}
}

type CacheEntry[T any] struct {
	Data       T     `json:"data"`
	Expiration int64 `json:"expiration"`
}

func SetCache[T any](key string, value T) error {
	entry := CacheEntry[T]{Data: value, Expiration: time.Now().Unix() + CacheDuration}
	Konfig.Set("cache."+key, entry)
	return nil
}

func GetCache[T any](key string) (T, error) {
	entry := CacheEntry[T]{}
	err := Konfig.Unmarshal("cache."+key, &entry)
	if err != nil {
		return entry.Data, err
	}
	if entry.Expiration < time.Now().Unix() {
		return entry.Data, ErrExpired
	}
	return entry.Data, nil
}

func ClearCache(key string) {
	Konfig.Delete("cache." + key)
}

func GetActive(key string) (string, error) {
	val := Konfig.Get("active." + key)
	if val == nil {
		return "", ErrNotFound
	}
	return val.(string), nil
}

func SetActive(key string, value string) {
	Konfig.Set("active."+key, value)
}

func ClearActive(key string) {
	Konfig.Delete("active." + key)
}

func Save() {
	b, _ := json.MarshalIndent(Konfig.Raw(), "", "\t")
	os.WriteFile(settingsFile, b, 0755)
}

func GetAuthToken() string {
	return Konfig.String("auth.token")
}

func SetAuthToken(token string) {
	Konfig.Set("auth.token", token)
}
