package ast

import "strings"

func RemoveStringQuote(target string) string {
	if strings.HasPrefix(target, "\"") || strings.HasPrefix(target, "'") {
		target = target[1:]
	}
	if strings.HasSuffix(target, "\"") || strings.HasSuffix(target, "'") {
		target = target[:len(target)-1]
	}
	return target
}
