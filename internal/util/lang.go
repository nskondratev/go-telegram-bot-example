package util

import "strings"

func Normalize(langCode string) (res string) {
	tokens := strings.Split(langCode, "-")
	if len(tokens) > 0 {
		res = tokens[0]
	}
	return res
}

func GetTargetLang(recognized, source, target string) string {
	norm := Normalize(recognized)
	if norm == target {
		return source
	} else {
		return target
	}
}
