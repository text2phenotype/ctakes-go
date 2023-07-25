package tokenizer

import (
	"text2phenotype.com/fdl/utils"
	"unicode"
)

type ContractionResult struct {
	WordTokenLen        int
	ContractionTokenLen int
}

type ContractionsPTB interface {
	IsContractionThatStartsWithApostrophe(currentPosition int, runes []rune) bool
	TokenLengthCheckingForSingleQuoteWordsToKeepTogether(runes []rune) int
	LenOfFirstTokenInContraction(runes []rune) int
	LenOfSecondTokenInContraction(runes []rune) int
	LenOfThirdTokenInContraction(runes []rune) int
	GetLengthIfNextApostIsMiddleOfContraction(currentPosition int, nextNonLetterOrNonDigit int, runes []rune) *ContractionResult
}

type contractionsPTBImpl struct {
	MultiTokenWordsLookup                        map[string]int
	MultiTokenWords                              []string
	MultiTokenWordLenToken1                      []int
	MultiTokenWordLenToken2                      []int
	MultiTokenWordLenToken3                      []int
	possibleContractionEndings                   [][]rune
	contractionsStartingWithApostrophe           [][]rune
	fullWordsNotToBreakAtApostrophe              [][]rune
	lettersAfterApostropheForMiddleOfContraction []rune
	hyphenatedPTB                                HyphenatedPTB
}

func (o *contractionsPTBImpl) IsContractionThatStartsWithApostrophe(currentPosition int, runes []rune) bool {
	r := runes[currentPosition:]
	for _, s := range o.contractionsStartingWithApostrophe {
		if startsWithWithoutBeingFollowedByLetter(r, s) {
			return true
		}
	}

	return false
}

func (o *contractionsPTBImpl) TokenLengthCheckingForSingleQuoteWordsToKeepTogether(runes []rune) int {

	firstBreak := utils.IndexOfRune(runes, apostrophe, 0)
	if firstBreak <= 0 {
		return -1
	}

	if firstBreak+1 == len(runes) {
		return firstBreak
	}

	secondBreak := findNextNonAlphaNum(runes, firstBreak+1)

	if o.breakAtApostrophe(runes, firstBreak) {
		return firstBreak
	}

	if secondBreak == len(runes) {
		return secondBreak
	}

	if runes[secondBreak] != hyphenOrMinusSign {
		return secondBreak
	} else {
		l := o.hyphenatedPTB.LenIfHyphenatedSuffix(runes, secondBreak)
		if l > 0 {
			return secondBreak + l
		}
		return secondBreak
	}
}

func (o *contractionsPTBImpl) LenOfFirstTokenInContraction(runes []rune) int {
	index, isOk := o.MultiTokenWordsLookup[string(runes)]
	if !isOk {
		return -1
	}

	return o.MultiTokenWordLenToken1[index]
}

func (o *contractionsPTBImpl) LenOfSecondTokenInContraction(runes []rune) int {
	index, isOk := o.MultiTokenWordsLookup[string(runes)]
	if !isOk {
		return -1
	}

	return o.MultiTokenWordLenToken2[index]
}

func (o *contractionsPTBImpl) LenOfThirdTokenInContraction(runes []rune) int {
	index, isOk := o.MultiTokenWordsLookup[string(runes)]
	if !isOk {
		return -1
	}

	return o.MultiTokenWordLenToken3[index]
}

func (o *contractionsPTBImpl) GetLengthIfNextApostIsMiddleOfContraction(position int, nextNonLetterDigit int, runes []rune) *ContractionResult {
	if position < 0 {
		return nil
	}

	if len(runes) < position+3 {
		return nil
	}

	apostrophePosition := utils.IndexOfRune(runes, apostrophe, position)

	if nextNonLetterDigit != apostrophePosition {
		return nil
	}

	if apostrophePosition < 1 || apostrophePosition >= len(runes)-1 || utils.RunesStartWith(runes, "n't") {
		return nil
	}

	letterAfterApostrophe := runes[apostrophePosition+1]
	if utils.IndexOfRune(o.lettersAfterApostropheForMiddleOfContraction, letterAfterApostrophe, 0) == -1 {
		return nil
	}

	subseqentNonAlphaNum := findNextNonAlphaNum(runes, apostrophePosition+1)
	restStartingWithApostrophe := runes[apostrophePosition:subseqentNonAlphaNum]

	prev := runes[apostrophePosition-1]

	negRunes := []rune("n't")
	for _, s := range o.possibleContractionEndings {
		lenAfterApostrophe := len(s) - 1
		isNeg := utils.EqualsRunes(s, negRunes)
		if isNeg {
			lenAfterApostrophe--
		}
		if len(runes) < apostrophePosition+lenAfterApostrophe {
			continue
		}

		if isNeg && prev == 'n' && runes[apostrophePosition+1] == 't' && len(runes) == apostrophePosition+1+1 {
			contractionResult := ContractionResult{
				ContractionTokenLen: 3,
				WordTokenLen:        apostrophePosition - 1 - position,
			}
			return &contractionResult
		} else if utils.EqualsRunes(restStartingWithApostrophe, s) {
			contractionResult := ContractionResult{
				ContractionTokenLen: len(s),
				WordTokenLen:        apostrophePosition - position,
			}
			return &contractionResult

		}

		if len(runes) == apostrophePosition+lenAfterApostrophe+1 {
			continue
		}

		var after rune
		if len(restStartingWithApostrophe) <= position+lenAfterApostrophe+1 {
			after = '\000'
		} else {
			after = restStartingWithApostrophe[position+lenAfterApostrophe+1]
		}

		if utils.RunesStartWith(restStartingWithApostrophe, string(s)) && unicode.IsLetter(prev) && !unicode.IsLetter(after) {
			contractionResult := ContractionResult{
				ContractionTokenLen: len(s),
				WordTokenLen:        apostrophePosition - position,
			}
			return &contractionResult
		} else if utils.EqualsRunes(s, negRunes) && prev == 'n' && utils.RunesStartWith(restStartingWithApostrophe, "'t") && !unicode.IsLetter(after) {
			contractionResult := ContractionResult{
				ContractionTokenLen: 3,
				WordTokenLen:        apostrophePosition - 1 - position,
			}
			return &contractionResult
		}
	}
	return nil
}

func (o *contractionsPTBImpl) breakAtApostrophe(runes []rune, positionOfApostropheToTest int) bool {

	if len(runes) == positionOfApostropheToTest+1 {
		return true
	}

	if positionOfApostropheToTest == 0 {
		return false
	}

	if allDigits(runes[0:positionOfApostropheToTest]) && runes[positionOfApostropheToTest+1] == 's' {
		if len(runes) < positionOfApostropheToTest+3 {
			return false
		}
		after := runes[positionOfApostropheToTest+2]
		if unicode.IsDigit(after) || unicode.IsLetter(after) {
			return true
		}
		return false
	} else {
		for _, comparison := range o.fullWordsNotToBreakAtApostrophe {
			if utils.EqualsRunes(comparison, runes) {
				return false
			}
		}
	}

	return true
}

func allDigits(runes []rune) bool {
	for i := 1; i < len(runes); i++ {
		if !unicode.IsDigit(runes[i]) {
			return false
		}
	}
	return true
}

func NewContractionsPTB(hyphenatedPTB HyphenatedPTB) ContractionsPTB {
	result := contractionsPTBImpl{
		MultiTokenWords:                    []string{"cannot", "gonna", "gotta", "lemme", "wanna", "whaddya", "whatcha"},
		MultiTokenWordsLookup:              make(map[string]int),
		hyphenatedPTB:                      hyphenatedPTB,
		MultiTokenWordLenToken1:            []int{3, 3, 3, 3, 3, 3, 3},
		MultiTokenWordLenToken2:            []int{3, 2, 2, 2, 2, 2, 1},
		MultiTokenWordLenToken3:            []int{0, 0, 0, 0, 0, 2, 3},
		possibleContractionEndings:         utils.StringsToRunes("'s", "'ve", "'re", "'ll", "'d", "'n", "n't"),
		contractionsStartingWithApostrophe: utils.StringsToRunes("'tis", "'twas"),
		fullWordsNotToBreakAtApostrophe:    utils.StringsToRunes("p'yongyang"),
	}

	for _, s := range result.possibleContractionEndings {
		indexLetterAfter := utils.IndexOfRune(s, apostrophe, 0) + 1
		if indexLetterAfter == 0 {
			continue
		}
		result.lettersAfterApostropheForMiddleOfContraction = append(result.lettersAfterApostropheForMiddleOfContraction, s[indexLetterAfter])

	}

	for i, s := range result.MultiTokenWords {
		result.MultiTokenWordsLookup[s] = i
	}

	return &result
}
