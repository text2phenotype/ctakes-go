package utils

import (
	"unicode/utf8"
)

func StringsToRunes(strs ...string) [][]rune {
	result := make([][]rune, len(strs))
	for i, s := range strs {
		result[i] = []rune(s)
	}
	return result
}

func IndexOfRune(runes []rune, c rune, fromIndex int) int {

	for i := fromIndex; i < len(runes); i++ {
		r := runes[i]
		if r == c {
			return i
		}
	}

	return -1
}

func IndexOfRunes(source []rune, target []rune) int {
	sIDx := 0
	tIDx := 0
	for sIDx < len(source) && tIDx < len(target) {
		if source[sIDx] == target[tIDx] {
			if sIDx == len(target)-1 {
				return sIDx - tIDx
			}

			tIDx++
		} else {
			tIDx = 0
		}

		sIDx++
	}

	return -1
}

func ContainRunes(source []rune, target []rune) bool {
	return IndexOfRunes(source, target) != -1
}

func EqualsRunes(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func RunesEndWith(runes []rune, s string) bool {

	if len(runes) < len(s) {
		return false
	}

	for i, c := range s {
		idx := len(runes) - len(s) + i
		if runes[idx] != c {
			return false
		}
	}
	return true
}

func RunesStartWith(runes []rune, s string) bool {

	if len(runes) < len(s) {
		return false
	}

	for i, c := range s {
		if runes[i] != c {
			return false
		}
	}
	return true
}

func MakeRuneByteSlices(txt string) ([]rune, []int) {
	runesCount := utf8.RuneCountInString(txt)
	runes := make([]rune, runesCount)
	bytes := make([]int, runesCount)

	bytesOffset := 0
	l := len(txt)
	for i := 0; i < runesCount && bytesOffset < l; i++ {
		ch, chSize := utf8.DecodeRuneInString(txt[bytesOffset:])
		runes[i] = ch
		bytes[i] = bytesOffset
		bytesOffset += chSize

	}
	return runes, bytes
}

func IsWhiteSpace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}
