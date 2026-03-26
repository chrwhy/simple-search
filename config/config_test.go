package config

import (
	"os"
	"runtime"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear all SS_ env vars to test defaults
	envVars := []string{
		"SS_DB_PATH", "SS_LIB_PATH", "SS_DATA_PATH",
		"SS_NAME_TABLE", "SS_CONTENT_TABLE",
		"SS_NAME_TOKENIZER", "SS_CONTENT_TOKENIZER",
	}
	for _, k := range envVars {
		os.Unsetenv(k)
	}

	cfg := Load()

	if cfg.DBPath != "example.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "example.db")
	}
	if cfg.DataPath != "data/sample.json" {
		t.Errorf("DataPath = %q, want %q", cfg.DataPath, "data/sample.json")
	}
	if cfg.NameTable != "docs_name" {
		t.Errorf("NameTable = %q, want %q", cfg.NameTable, "docs_name")
	}
	if cfg.ContentTable != "docs_content" {
		t.Errorf("ContentTable = %q, want %q", cfg.ContentTable, "docs_content")
	}
	if cfg.NameTokenizer != "simple" {
		t.Errorf("NameTokenizer = %q, want %q", cfg.NameTokenizer, "simple")
	}
	if cfg.ContentTokenizer != "simple 0" {
		t.Errorf("ContentTokenizer = %q, want %q", cfg.ContentTokenizer, "simple 0")
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("SS_DB_PATH", "/tmp/test.db")
	os.Setenv("SS_NAME_TABLE", "my_names")
	os.Setenv("SS_CONTENT_TABLE", "my_contents")
	defer func() {
		os.Unsetenv("SS_DB_PATH")
		os.Unsetenv("SS_NAME_TABLE")
		os.Unsetenv("SS_CONTENT_TABLE")
	}()

	cfg := Load()

	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("DBPath = %q, want %q", cfg.DBPath, "/tmp/test.db")
	}
	if cfg.NameTable != "my_names" {
		t.Errorf("NameTable = %q, want %q", cfg.NameTable, "my_names")
	}
	if cfg.ContentTable != "my_contents" {
		t.Errorf("ContentTable = %q, want %q", cfg.ContentTable, "my_contents")
	}
}

func TestDetectLibPath(t *testing.T) {
	path := detectLibPath()
	expected := "./lib/" + runtime.GOOS + "-" + runtime.GOARCH + "/libsimple"
	if path != expected {
		t.Errorf("detectLibPath() = %q, want %q", path, expected)
	}
}

func TestValidateTokenizer(t *testing.T) {
	tests := []struct {
		name    string
		tok     string
		wantErr bool
	}{
		{"simple", "simple", false},
		{"simple with args", "simple 0", false},
		{"unicode61", "unicode61", false},
		{"trigram", "trigram", false},
		{"unknown", "foobar", true},
		{"empty", "", true},
		{"injection attempt", "simple'; DROP TABLE--", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTokenizer(tt.tok)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTokenizer(%q) error = %v, wantErr %v", tt.tok, err, tt.wantErr)
			}
		})
	}
}
