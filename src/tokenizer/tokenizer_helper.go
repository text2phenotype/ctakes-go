package tokenizer

import (
	"text2phenotype.com/fdl/utils"
	"unicode"
)

func findNextNonAlphaNum(runes []rune, fromIndex int) int {
	for i := fromIndex; i < len(runes); i++ {
		if !(unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i])) {
			return i
		}
	}
	return len(runes)
}

func startsWithWithoutBeingFollowedByLetter(s []rune, compareTo []rune) bool {

	if utils.RunesStartWith(s, string(compareTo)) {
		if len(s) == len(compareTo) {
			return true
		}
		next := s[len(compareTo)]

		return !unicode.IsLetter(next)
	}
	return false
}
