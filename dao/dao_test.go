package dao

import "testing"

func TestSanitizeMatch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"it's", "it''s"},
		{"a'b'c", "a''b''c"},
		{"no quotes", "no quotes"},
		{"", ""},
		{"'", "''"},
	}
	for _, tt := range tests {
		got := sanitizeMatch(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeMatch(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
