package tokenizer

import (
	"text2phenotype.com/fdl/utils"
	"unicode"
)

type HyphenatedPTB interface {
	TokenLengthCheckingForHyphenatedTerms(runes []rune) int
	LenIfHyphenatedSuffix(runes []rune, secondBreak int) int
}

type hyphenatedPTBImpl struct {
	hyphenatedPrefixesLookup map[string]bool
	hyphenatedSuffixesLookup map[string]bool
	hyphenatedWordsLookup    map[string]bool
}

func (o *hyphenatedPTBImpl) TokenLengthCheckingForHyphenatedTerms(runes []rune) int {

	firstBreak := utils.IndexOfRune(runes, hyphenOrMinusSign, 0)
	if firstBreak <= 0 {
		return -1
	}

	if firstBreak+1 == len(runes) {
		return firstBreak
	}

	secondBreak := findNextNonAlphaNum(runes, firstBreak+1)
	thirdBreak := -1
	if secondBreak != len(runes) {
		thirdBreak = findNextNonAlphaNum(runes, secondBreak+1)
	}

	if secondBreak == len(runes) {
		return o.lenIncludingHyphensToKeep(runes, firstBreak, 1, secondBreak, thirdBreak)
	} else if runes[secondBreak] == hyphenOrMinusSign {
		return o.lenIncludingHyphensToKeep(runes, firstBreak, 2, secondBreak, thirdBreak)
	} else if runes[secondBreak] == apostrophe {
		return o.lenIncludingHyphensToKeep(runes, firstBreak, 1, secondBreak, thirdBreak)
	} else if unicode.IsSpace(runes[secondBreak]) {
		return o.lenIncludingHyphensToKeep(runes, firstBreak, 1, secondBreak, thirdBreak)
	} else {
		return o.lenIncludingHyphensToKeep(runes, firstBreak, 1, secondBreak, thirdBreak)
	}

}

func (o *hyphenatedPTBImpl) LenIfHyphenatedSuffix(runes []rune, position int) int {
	next := findNextNonAlphaNum(runes, position+1)
	possibleSuffix := runes[position:next]
	if utils.RunesStartWith(runes[position:], "-o-") {
		next = findNextNonAlphaNum(runes, position+3)
		possibleSuffix = runes[position:next]
	}
	lookup := false

	_, lookup = o.hyphenatedSuffixesLookup[string(possibleSuffix)]

	if lookup {
		return len(possibleSuffix)
	}
	return -1
}

func (o *hyphenatedPTBImpl) lenIncludingHyphensToKeep(runes []rune, indexOfFirstHyphen int, numberOfHyphensToConsiderKeeping int, secondBreak int, thirdBreak int) int {

	var possibleSuffix []rune
	var lookup bool

	if numberOfHyphensToConsiderKeeping > 2 || numberOfHyphensToConsiderKeeping < 1 {
		return -1
	}

	if numberOfHyphensToConsiderKeeping == 2 {
		possibleSuffix = runes[indexOfFirstHyphen:thirdBreak]
		lookup = o.hyphenatedSuffixesLookup[string(possibleSuffix)]
		if lookup {
			return thirdBreak
		}
	}

	possibleSuffix = runes[indexOfFirstHyphen:secondBreak]
	lookup = o.hyphenatedSuffixesLookup[string(possibleSuffix)]
	if lookup {
		if thirdBreak > secondBreak {
			possibleSuffix = runes[secondBreak:thirdBreak]
			lookup = o.hyphenatedSuffixesLookup[string(possibleSuffix)]
			if lookup {
				return thirdBreak
			}
		}
		return secondBreak
	}

	if numberOfHyphensToConsiderKeeping > 1 {
		possibleHyphenatedWordsLookupMatch := string(runes[:secondBreak])
		possibleSuffix = runes[secondBreak:thirdBreak]
		lookup = o.hyphenatedWordsLookup[possibleHyphenatedWordsLookupMatch] && o.hyphenatedSuffixesLookup[string(possibleSuffix)]
		if lookup {
			return thirdBreak
		}
	}

	possiblePrefix := runes[:indexOfFirstHyphen+1]

	lookup = o.hyphenatedPrefixesLookup[string(possiblePrefix)]

	if lookup && numberOfHyphensToConsiderKeeping > 1 {
		possibleHyphenatedWordsLookupMatch := string(runes[indexOfFirstHyphen+1 : thirdBreak])
		lookup2 := o.hyphenatedWordsLookup[possibleHyphenatedWordsLookupMatch]
		if lookup2 {
			return thirdBreak
		}
	}

	if numberOfHyphensToConsiderKeeping == 1 {
		if lookup {
			return secondBreak
		}
	}

	if numberOfHyphensToConsiderKeeping == 2 {
		if lookup {
			possibleSuffix = runes[secondBreak:thirdBreak]
			lookup2 := o.hyphenatedSuffixesLookup[string(possibleSuffix)]
			if lookup2 {
				return thirdBreak
			}
			return secondBreak
		}
	}
	possibleHyphenatedWordsLookupMatch := string(runes[:secondBreak])
	lookup = o.hyphenatedWordsLookup[possibleHyphenatedWordsLookupMatch]
	if lookup {
		return secondBreak
	}

	return indexOfFirstHyphen
}

func createHyphenatedSuffixes() map[string]bool {
	return map[string]bool{
		"-esque":    true,
		"-ette":     true,
		"-fest":     true,
		"-fold":     true,
		"-gate":     true,
		"-itis":     true,
		"-less":     true,
		"-most":     true,
		"-o-torium": true,
		"-rama":     true,
		"-wise":     true,
	}
}

func createHyphenatedPrefixes() map[string]bool {
	return map[string]bool{
		"e-":       true,
		"a-":       true,
		"u-":       true,
		"x-":       true,
		"agro-":    true,
		"ante-":    true,
		"anti-":    true,
		"arch-":    true,
		"be-":      true,
		"bi-":      true,
		"bio-":     true,
		"co-":      true,
		"counter-": true,
		"cross-":   true,
		"cyber-":   true,
		"de-":      true,
		"eco-":     true,
		"ex-":      true,
		"extra-":   true,
		"inter-":   true,
		"intra-":   true,
		"macro-":   true,
		"mega-":    true,
		"micro-":   true,
		"mid-":     true,
		"mini-":    true,
		"multi-":   true,
		"neo-":     true,
		"non-":     true,
		"over-":    true,
		"pan-":     true,
		"para-":    true,
		"peri-":    true,
		"post-":    true,
		"pre-":     true,
		"pro-":     true,
		"pseudo-":  true,
		"quasi-":   true,
		"re-":      true,
		"semi-":    true,
		"sub-":     true,
		"super-":   true,
		"tri-":     true,
		"ultra-":   true,
		"un-":      true,
		"uni-":     true,
		"vice-":    true,
		"electro-": true,
		"gasto-":   true,
		"homo-":    true,
		"hetero-":  true,
		"ortho-":   true,
		"phospho-": true,
	}
}

func createHyphenatedWordsLookup() map[string]bool {
	return map[string]bool{
		"mm-hm":  true,
		"mm-mm":  true,
		"o-kay":  true,
		"uh-huh": true,
		"uh-oh":  true,
	}
}

func NewHyphenatedPTB() HyphenatedPTB {
	result := hyphenatedPTBImpl{
		hyphenatedPrefixesLookup: createHyphenatedPrefixes(),
		hyphenatedSuffixesLookup: createHyphenatedSuffixes(),
		hyphenatedWordsLookup:    createHyphenatedWordsLookup(),
	}

	return &result
}
