package config

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath           string
	LibPath          string
	DataPath         string
	NameTable        string
	ContentTable     string
	NameTokenizer    string
	ContentTokenizer string
}

// knownTokenizers lists tokenizer prefixes that are safe to embed in CREATE TABLE.
var knownTokenizers = map[string]bool{
	"simple":   true,
	"unicode61": true,
	"porter":   true,
	"ascii":    true,
	"trigram":  true,
}

func Load() *Config {
	// Load .env file if present (silently ignore if missing)
	_ = godotenv.Load()

	return &Config{
		DBPath:           getEnv("SS_DB_PATH", "example.db"),
		LibPath:          getEnv("SS_LIB_PATH", detectLibPath()),
		DataPath:         getEnv("SS_DATA_PATH", "data/sample.json"),
		NameTable:        getEnv("SS_NAME_TABLE", "docs_name"),
		ContentTable:     getEnv("SS_CONTENT_TABLE", "docs_content"),
		NameTokenizer:    getEnv("SS_NAME_TOKENIZER", "simple"),
		ContentTokenizer: getEnv("SS_CONTENT_TOKENIZER", "simple 0"),
	}
}

// Validate checks that config values are sane. Returns an error describing the
// first problem found, or nil if everything is valid.
func (c *Config) Validate() error {
	if _, err := os.Stat(c.LibPath); err != nil {
		return fmt.Errorf("libsimple 库文件不存在: %s", c.LibPath)
	}
	if _, err := os.Stat(c.DataPath); err != nil {
		return fmt.Errorf("数据文件不存在: %s", c.DataPath)
	}
	if err := validateTokenizer(c.NameTokenizer); err != nil {
		return fmt.Errorf("NameTokenizer: %w", err)
	}
	if err := validateTokenizer(c.ContentTokenizer); err != nil {
		return fmt.Errorf("ContentTokenizer: %w", err)
	}
	return nil
}

func validateTokenizer(tok string) error {
	// Tokenizer format: "name [args...]", e.g. "simple 0"
	base, _, _ := strings.Cut(tok, " ")
	if !knownTokenizers[base] {
		return fmt.Errorf("未知分词器 %q; 已知: simple, unicode61, porter, ascii, trigram", base)
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func detectLibPath() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return fmt.Sprintf("./lib/%s-%s/libsimple", goos, goarch)
}
