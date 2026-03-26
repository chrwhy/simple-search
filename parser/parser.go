package parser

import (
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenChinese tokenType = iota
	tokenAlpha
	tokenDigit
	tokenMixed
	tokenPunct
)

func classifyToken(s string) tokenType {
	hasChinese := false
	hasAlpha := false
	hasDigit := false

	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			hasChinese = true
		case unicode.IsLetter(r):
			hasAlpha = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	mixed := 0
	if hasChinese {
		mixed++
	}
	if hasAlpha {
		mixed++
	}
	if hasDigit {
		mixed++
	}
	if mixed > 1 {
		return tokenMixed
	}
	if hasChinese {
		return tokenChinese
	}
	if hasAlpha {
		return tokenAlpha
	}
	if hasDigit {
		return tokenDigit
	}
	return tokenPunct
}

// splitMixed 将中英数混合的 token 按字符类型拆分
func splitMixed(s string) []string {
	var result []string
	var buf strings.Builder
	var lastType rune

	for _, r := range s {
		var t rune
		switch {
		case unicode.Is(unicode.Han, r):
			t = 'C'
		case unicode.IsLetter(r):
			t = 'E'
		case unicode.IsDigit(r):
			t = 'D'
		default:
			t = 'O'
		}

		if lastType != 0 && t != lastType {
			if buf.Len() > 0 {
				result = append(result, buf.String())
				buf.Reset()
			}
		}
		buf.WriteRune(r)
		lastType = t
	}
	if buf.Len() > 0 {
		result = append(result, buf.String())
	}
	return result
}

func escapeQuote(s string) string {
	s = strings.ReplaceAll(s, `"`, `""`)
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}
