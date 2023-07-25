package lookup

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func NewTermTokenizer() (func(term string) []string, error) {

	splitPattern, err := regexp.Compile(`\s+`)
	if err != nil {
		return nil, err
	}

	suffixes := getSuffixes()
	prefixes := getPrefixes()

	return func(term string) []string {
		if len(term) == 0 {
			return []string{}
		}

		splits := splitPattern.Split(term, -1)
		if len(splits) == 0 {
			return []string{}
		}

		sb := make([]string, 0, len(splits))
		for _, split := range splits {
			tokens := getTokens(split, prefixes, suffixes)
			sb = append(sb, tokens...)
		}

		return sb
	}, nil
}

func getTokens(word string, prefixes map[string]bool, suffixes map[string]bool) []string {
	var tokens []string
	var sb strings.Builder
	offset := 0
	for offset < len(word) {
		c, size := utf8.DecodeRuneInString(word[offset:])
		offset += size

		if unicode.IsLetter(c) || unicode.IsDigit(c) {
			sb.WriteRune(c)
			continue
		}

		if c == '-' && (isPrefix(sb.String(), prefixes) || isSuffix(word, offset, suffixes)) {
			sb.WriteRune(c)
			continue
		}

		if (c == '\'' && isOwnerApostrophe(word, offset)) || (c == '.' && isNumberDecimal(word, offset)) {
			if sb.Len() > 0 {
				tokens = append(tokens, sb.String())
				sb.Reset()
			}

			sb.WriteRune(c)
			continue
		}

		if sb.Len() != 0 {
			tokens = append(tokens, sb.String())
			sb.Reset()
		}

		tokens = append(tokens, string(c))
	}

	if sb.Len() != 0 {
		tokens = append(tokens, sb.String())
	}
	return tokens
}

func isPrefix(word string, prefixes map[string]bool) bool {
	return prefixes[word]
}

func isSuffix(word string, startOffset int, suffixes map[string]bool) bool {
	if len(word) <= startOffset {
		return false
	}
	nextCharTerm := getNextCharTerm(word[startOffset:])
	if len(nextCharTerm) == 0 {
		return false
	}
	return suffixes[nextCharTerm]
}

func isOwnerApostrophe(word string, startOffset int) bool {
	c, size := utf8.DecodeRuneInString(word[startOffset:])
	return len(word) <= startOffset+size && c == 's'
}

func isNumberDecimal(word string, startOffset int) bool {
	c, size := utf8.DecodeRuneInString(word[startOffset:])
	return len(word) <= startOffset+size && unicode.IsDigit(c)
}

func getNextCharTerm(word string) string {
	for offset, c := range word {
		if !(unicode.IsLetter(c) || unicode.IsDigit(c)) {
			return word[:offset]
		}
	}
	return word
}

func getPrefixes() map[string]bool {
	return map[string]bool{
		"e":       true,
		"a":       true,
		"u":       true,
		"x":       true,
		"agro":    true,
		"ante":    true,
		"anti":    true,
		"arch":    true,
		"be":      true,
		"bi":      true,
		"bio":     true,
		"co":      true,
		"counter": true,
		"cross":   true,
		"cyber":   true,
		"de":      true,
		"eco":     true,
		"ex":      true,
		"extra":   true,
		"inter":   true,
		"intra":   true,
		"macro":   true,
		"mega":    true,
		"micro":   true,
		"mid":     true,
		"mini":    true,
		"multi":   true,
		"neo":     true,
		"non":     true,
		"over":    true,
		"pan":     true,
		"para":    true,
		"peri":    true,
		"post":    true,
		"pre":     true,
		"pro":     true,
		"pseudo":  true,
		"quasi":   true,
		"re":      true,
		"semi":    true,
		"sub":     true,
		"super":   true,
		"tri":     true,
		"ultra":   true,
		"un":      true,
		"uni":     true,
		"vice":    true,
		"electro": true,
		"gasto":   true,
		"homo":    true,
		"hetero":  true,
		"ortho":   true,
		"phospho": true,
	}
}

func getSuffixes() map[string]bool {
	return map[string]bool{
		"esque":    true,
		"ette":     true,
		"fest":     true,
		"fold":     true,
		"gate":     true,
		"itis":     true,
		"less":     true,
		"most":     true,
		"o-torium": true,
		"rama":     true,
		"wise":     true,
	}
}
