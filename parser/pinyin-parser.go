package parser

import (
	"github.com/chrwhy/open-pinyin/dict"
	pinyin "github.com/chrwhy/open-pinyin/parser"
	"github.com/chrwhy/open-pinyin/util"
)

const (
	SubPinyinStopSign = 3
)

func ParsePinyinClause(input string) string {
	pinyinGroups := pinyin.Parse(input)
	pinyinInitial := pinyin.ParseInitial(input)
	if len(pinyinInitial) > 0 {
		pinyinGroups = append(pinyinGroups, pinyinInitial)
	}
	clause := ""
	for i, pinyinGroup := range pinyinGroups {
		for j, _ := range pinyinGroup {
			if _, ok := dict.SUB_PINYIN[pinyinGroup[j]]; ok {
				if j != len(pinyinGroup)-1 && len(pinyinGroup[j]) > 1 {
					pinyinGroup[j] = "\"" + pinyinGroup[j] + string(rune(SubPinyinStopSign)) + "\""
				}
			}
		}
		clause += util.Concat(pinyinGroup, "+")
		if len(pinyinGroups) > 1 && i != len(pinyinGroups)-1 {
			clause += " OR "
		}
	}
	return clause
}
