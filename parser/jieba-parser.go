package parser

import (
	"strings"

	"github.com/yanyiwu/gojieba"
)

var jieba *gojieba.Jieba

func InitJieba() {
	jieba = gojieba.NewJieba()
}

func FreeJieba() {
	jieba.Free()
}

// ParseJiebaClause 使用 jieba 分词将用户输入转换为 FTS5 MATCH 条件。
//
// 规则：
//   - 中文词组保持整词，用双引号包裹
//   - 英文按空格拆分，转小写，有拼音候选时用 (pinyin OR token) 组合，否则 ("token")
//   - 数字用双引号包裹，精确匹配
//   - 所有 token 之间用 AND 连接
//
// 示例:
//
//	"我爱China"  → `"我" AND "爱" AND (chi+na OR china)`
//	"周杰伦 Jay Chou" → `"周杰伦" AND (j+a+y OR j+a+y OR jay) AND (chou OR chou)`
func ParseJiebaClause(query string) string {
	words := jieba.Cut(query, true)

	var parts []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		cat := classifyToken(word)
		switch cat {
		case tokenChinese:
			word = escapeQuote(word)
			parts = append(parts, `"`+word+`"`)
		case tokenAlpha:
			lower := strings.ToLower(word)
			pinyinClause := ParsePinyinClause(lower)
			if len(pinyinClause) > 0 {
				parts = append(parts, "("+pinyinClause+" OR "+lower+")")
			} else {
				parts = append(parts, `("`+lower+`")`)
			}
		case tokenDigit:
			parts = append(parts, `"`+word+`"`)
		case tokenMixed:
			subTokens := splitMixed(word)
			for _, st := range subTokens {
				sub := ParseJiebaClause(st)
				if sub != "" {
					parts = append(parts, sub)
				}
			}
		case tokenPunct:
			word = escapeQuote(word)
			parts = append(parts, `"`+word+`"`)
		}
	}

	return strings.Join(parts, " AND ")
}
