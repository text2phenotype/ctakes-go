package lemmatizer

import "strings"

func AnyOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

func StartsWithAny(s string, values ...string) bool {
	for _, v := range values {
		if strings.HasPrefix(s, v) {
			return true
		}
	}
	return false
}
