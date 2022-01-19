package strings2

import (
	"unicode"
)

// UpperFirst 首字母大写
func UpperFirst(str string) string {
	ret := []rune(str)
	if len(ret) > 0 {
		ret[0] = unicode.ToUpper(ret[0])
	}
	return string(ret)
}

// LowerFirst 首字母小写
func LowerFirst(str string) string {
	ret := []rune(str)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}
