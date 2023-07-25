package tokenizer

import (
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	period            = '.'
	comma             = ','
	newline           = '\n'
	cr                = '\r'
	hyphenOrMinusSign = '-'
	apostrophe        = '\''
	dash              = '-'
	at                = '@'
)

func NewTokenizerPTB() func(sent *types.Sentence) error {

	hyphenatedPTB := NewHyphenatedPTB()

	contractionsPTB := NewContractionsPTB(hyphenatedPTB)

	return func(sent *types.Sentence) error {
		return tokenizeSentence(sent, contractionsPTB, hyphenatedPTB)
	}
}

func createToken(txt string, begin int32, end int32, IsWord bool, IsNumber bool, IsSymbol bool, IsNewline bool, IsPunct bool) *types.Token {
	newToken := types.Token{
		Span: types.Span{
			Begin: begin,
			End:   end,
			Text:  utils.GlobalStringStore().GetPointer(txt),
		},
		IsPunct:   IsPunct,
		IsWord:    IsWord,
		IsNumber:  IsNumber,
		IsSymbol:  IsSymbol,
		IsNewline: IsNewline,
		Shape:     types.GetShape(txt),
	}

	return &newToken
}

func tokenizeSentence(sent *types.Sentence, contractionsPTB ContractionsPTB, hyphenatedPTB HyphenatedPTB) error {
	if sent.Text == nil || len(*sent.Text) == 0 {
		return nil
	}

	txt := strings.ToLower(*sent.Text)
	//txt := string([]rune{'Ѡ', 'Ѡ', '\n', '\r'})
	//txt := string([]rune{'Ѡ', 'Ѡ', '\r', '\n'})
	//sent.Text = &txt
	runes, bytes := utils.MakeRuneByteSlices(txt)

	runesLen := len(runes)

	currentPosition := 0

	currentPosition = findFirstCharOfNextToken(currentPosition, runes)
	if currentPosition < 0 {
		return nil
	}

	const NotSetIndicator = -999
	for currentPosition >= 0 {

		IsPunct := false
		IsWord := false
		IsSymbol := false
		IsNumber := false
		IsNewline := false

		firstCharOfToken := runes[currentPosition]
		tokenLen := NotSetIndicator

		if currentPosition+1 >= runesLen {
			tokenLen = 1
			IsSymbol = true
		} else {
			nextChar := runes[currentPosition+1]
			if utils.IsWhiteSpace(nextChar) {
				tokenLen = 1
				IsPunct = unicode.IsPunct(firstCharOfToken)
				IsWord = unicode.IsLetter(firstCharOfToken)
				IsNumber = unicode.IsDigit(firstCharOfToken)
				IsNewline = firstCharOfToken == newline || firstCharOfToken == cr
				IsSymbol = !IsPunct && !IsWord && !IsNumber && !IsNewline
			} else {
				if firstCharOfToken == newline {
					tokenLen = 1
					IsNewline = true
				} else {
					if firstCharOfToken == cr {
						if nextChar != newline {
							tokenLen = 1
						} else {
							tokenLen = 2
						}
						IsNewline = true
					} else {
						if firstCharOfToken == period {
							ln := getLengthIfIsNumberThatStartsWithPeriod(currentPosition, runes)
							if ln > 0 {
								tokenLen = ln
								IsNumber = true
							} else if isEllipsis(bytes[currentPosition], txt) {
								tokenLen = 3
								IsPunct = true
							} else {
								tokenLen = 1
								IsPunct = true
							}
						} else if firstCharOfToken == hyphenOrMinusSign {
							tokenLen = 1
							IsPunct = true
						} else if firstCharOfToken == apostrophe {
							if contractionsPTB.IsContractionThatStartsWithApostrophe(currentPosition, runes) {
								tokenLen = 2
								IsWord = true
							} else {
								tokenLen = 1
								IsPunct = true
							}
						} else if unicode.IsPunct(firstCharOfToken) {
							tokenLen = 1
							IsPunct = true
						} else if unicode.IsLetter(firstCharOfToken) || unicode.IsDigit(firstCharOfToken) {
							obviouslyIsWord := true
							obviouslyIsNumber := true
							nextWhitespaceOrEndOfSegment := -1
							nextNonLetterOrNonDigit := -1
							nextNonLetterDigitApostrophe := -1
							nextNonTelephoneOrPostalChar := -1
							nextNonNumericChar := -1
							nextNonDigit := -1

							i := currentPosition
							ch := runes[i]

							for i < runesLen && !utils.IsWhiteSpace(ch) {

								ch = runes[i]

								charIsLetter := unicode.IsLetter(ch)
								charIsDigit := unicode.IsDigit(ch)

								if utils.IsWhiteSpace(ch) {
									if nextNonLetterOrNonDigit < 0 {
										nextNonLetterOrNonDigit = i
									}

									if nextNonLetterDigitApostrophe < 0 {
										nextNonLetterDigitApostrophe = i
									}

									if nextNonDigit < 0 {
										nextNonDigit = i
									}

									if nextNonTelephoneOrPostalChar < 0 {
										nextNonTelephoneOrPostalChar = i
									}
									if nextNonNumericChar < 0 {
										nextNonNumericChar = i
									}
									nextWhitespaceOrEndOfSegment = i
								} else if !(charIsLetter || charIsDigit) {
									obviouslyIsWord = false
									obviouslyIsNumber = false
									if nextNonLetterOrNonDigit < 0 {
										nextNonLetterOrNonDigit = i
									}

									if nextNonLetterDigitApostrophe < 0 && ch != apostrophe {
										nextNonLetterDigitApostrophe = i
									}

									if nextNonDigit < 0 {
										nextNonDigit = i
									}

									if nextNonTelephoneOrPostalChar < 0 && !isTelephoneNumberChar(ch) {
										nextNonTelephoneOrPostalChar = i
									}

									if nextNonNumericChar < 0 && !isNumericChar(ch) {
										nextNonNumericChar = i
									}
								} else if !charIsDigit {
									obviouslyIsNumber = false
									if nextNonDigit < 0 {
										nextNonDigit = i
									}
									if nextNonTelephoneOrPostalChar < 0 && !isTelephoneNumberChar(ch) {
										nextNonTelephoneOrPostalChar = i
									}
									if nextNonNumericChar < 0 && !isNumericChar(ch) {
										nextNonNumericChar = i
									}
								}

								i++
							}

							if i >= runesLen {
								if nextWhitespaceOrEndOfSegment < 0 {
									nextWhitespaceOrEndOfSegment = runesLen
								}
								if nextNonLetterOrNonDigit < 0 {
									nextNonLetterOrNonDigit = runesLen
								}
								if nextNonLetterDigitApostrophe < 0 {
									nextNonLetterDigitApostrophe = runesLen
								}
								if nextNonTelephoneOrPostalChar < 0 {
									nextNonTelephoneOrPostalChar = runesLen
								}
								if nextNonNumericChar < 0 {
									nextNonNumericChar = runesLen
								}
							}

							if obviouslyIsNumber {
								tokenLen = nextWhitespaceOrEndOfSegment - currentPosition
								IsNumber = true
							} else if obviouslyIsWord {
								substringRunes := runes[currentPosition:nextWhitespaceOrEndOfSegment]

								l := contractionsPTB.LenOfFirstTokenInContraction(substringRunes)
								if l > 0 {
									tokenLen = l
									IsWord = true

									var txt string
									if currentPosition+tokenLen > runesLen-1 {
										txt = (*sent.Text)[bytes[currentPosition]:]
									} else {
										txt = (*sent.Text)[bytes[currentPosition]:bytes[currentPosition+tokenLen]]
									}
									begin := int32(currentPosition) + sent.Begin
									end := begin + int32(tokenLen)
									newToken := createToken(txt, begin, end, IsWord, IsNumber, IsSymbol, IsNewline, IsPunct)
									sent.Tokens = append(sent.Tokens, newToken)
									currentPosition += tokenLen

									l = contractionsPTB.LenOfSecondTokenInContraction(substringRunes)

									tokenLen = l

									l = contractionsPTB.LenOfThirdTokenInContraction(substringRunes)
									if l > 0 {
										var txt string
										if currentPosition+tokenLen > runesLen-1 {
											txt = (*sent.Text)[bytes[currentPosition]:]
										} else {
											txt = (*sent.Text)[bytes[currentPosition]:bytes[currentPosition+tokenLen]]
										}
										begin := int32(currentPosition) + sent.Begin
										end := begin + int32(tokenLen)
										newToken := createToken(txt, begin, end, IsWord, IsNumber, IsSymbol, IsNewline, IsPunct)
										sent.Tokens = append(sent.Tokens, newToken)
										currentPosition += tokenLen

										tokenLen = l
									}
								} else {
									tokenLen = nextWhitespaceOrEndOfSegment - currentPosition
									IsWord = true
								}
							} else {

								var l int
								var cr *ContractionResult

								if nextNonLetterOrNonDigit < runesLen && runes[nextNonLetterOrNonDigit] == apostrophe {
									substringRunes := runes[currentPosition:nextWhitespaceOrEndOfSegment]

									l = contractionsPTB.TokenLengthCheckingForSingleQuoteWordsToKeepTogether(substringRunes)
									if l > nextNonLetterOrNonDigit-currentPosition {
										tokenLen = l

										IsNumber, IsWord = wordTokenOrNumToken(runes, currentPosition, tokenLen)
									}
								}
								if tokenLen == NotSetIndicator {
									cr = contractionsPTB.GetLengthIfNextApostIsMiddleOfContraction(currentPosition, nextNonLetterOrNonDigit, runes)
									if cr != nil {
										l = cr.WordTokenLen
										tokenLen = l
										IsWord = true
										c := runes[currentPosition+l]
										if c == 'n' || c == apostrophe {
											if tokenLen < 0 {
												errText := fmt.Sprintf("c = %q tokenLen = %d currentPosition = %d", c, tokenLen, currentPosition)
												return errors.New(errText)
											}

											if tokenLen > 0 {
												endOffsetIdx := currentPosition + tokenLen
												if endOffsetIdx >= len(bytes) {
													endOffsetIdx = -1
												}

												var txt string
												if endOffsetIdx < 0 || bytes[endOffsetIdx] >= len(*sent.Text) {
													txt = (*sent.Text)[bytes[currentPosition]:]
												} else {
													endOffset := bytes[endOffsetIdx]
													txt = (*sent.Text)[bytes[currentPosition]:endOffset]
												}
												begin := int32(currentPosition) + sent.Begin
												end := begin + int32(tokenLen)

												newToken := createToken(txt, begin, end, IsWord, IsNumber, IsSymbol, IsNewline, IsPunct)
												sent.Tokens = append(sent.Tokens, newToken)
												currentPosition += tokenLen // currentPosition
											}
											tokenLen = cr.ContractionTokenLen
											IsWord = true
										} else {
											errText := fmt.Sprintf("ERROR: getLengthIfNextApostIsMiddleOfContraction returned %d but the character %q after that is not 'n' or apostrophe ", l, c)
											return errors.New(errText)
										}

									} else if l = lenIfIsTelephoneNumber(currentPosition, runes, nextNonTelephoneOrPostalChar); l > 0 {
										tokenLen = l
										IsWord = true
									} else if l = lenIfIsPostalCode(currentPosition, runes, nextNonTelephoneOrPostalChar); l > 0 {
										tokenLen = l
										IsWord = true
									} else if l = lenIfIsUrl(currentPosition, runes, nextWhitespaceOrEndOfSegment); l > 0 {
										tokenLen = l
										IsWord = true
									} else if l = lenIfIsEmailAddress(currentPosition, runes, nextWhitespaceOrEndOfSegment); l > 0 {
										tokenLen = l
										IsWord = true
									} else if l = lenIfIsAbbreviation(currentPosition, runes, nextWhitespaceOrEndOfSegment); l > 0 {
										tokenLen = l
										IsWord = true
									} else {
										if nextNonLetterOrNonDigit < runesLen && runes[nextNonLetterOrNonDigit] == hyphenOrMinusSign {
											substringRunes := runes[currentPosition:nextWhitespaceOrEndOfSegment]
											l = hyphenatedPTB.TokenLengthCheckingForHyphenatedTerms(substringRunes)
											tokenLen = l
											if tokenLen < 0 {
												errText := fmt.Sprintf("Token len is negative: tokenLen = %d  currentPosition = %d nextNonLetterOrNonDigit = %d", tokenLen, currentPosition, nextNonLetterOrNonDigit)
												return errors.New(errText)
											}

											IsNumber, IsWord = wordTokenOrNumToken(runes, currentPosition, tokenLen)
										} else if l = lenIfIsNumberContainingComma(currentPosition, runes, nextNonNumericChar); (nextNonNumericChar > 0) && l > 0 {
											tokenLen = l
											IsNumber = true
										} else if nextNonLetterDigitApostrophe < runesLen && runes[nextNonLetterDigitApostrophe] == period {
											if nextNonDigit == runesLen-1 {
												tokenLen = nextNonDigit - currentPosition
												IsNumber = true
											} else if nextNonLetterDigitApostrophe == nextNonDigit {
												tokenLen = nextNonDigit + 1 + getLenToNextNonDigit(runes, nextNonDigit+1) - currentPosition
												IsNumber = true
											} else {
												tokenLen = nextNonLetterOrNonDigit - currentPosition
												IsNumber, IsWord = wordTokenOrNumToken(runes, currentPosition, tokenLen)
											}
										} else {
											tokenLen = nextNonLetterOrNonDigit - currentPosition
											IsNumber, IsWord = wordTokenOrNumToken(runes, currentPosition, tokenLen)
										}
									}

								}

							}

						} else {
							tokenLen = 1
							IsSymbol = true
						}
					}
				}
			}
		}

		if tokenLen < 0 {
			errText := fmt.Sprintf("Token len is negative: tokenLen = %d  currentPosition = %d", tokenLen, currentPosition)
			return errors.New(errText)
		}

		if tokenLen > 0 {

			endOffsetIdx := currentPosition + tokenLen
			if endOffsetIdx >= len(bytes) {
				endOffsetIdx = -1
			}

			var txt string
			if endOffsetIdx < 0 || bytes[endOffsetIdx] >= len(*sent.Text) {
				if bytes[currentPosition] >= len(*sent.Text) {
					errText := fmt.Sprintf("Tokenizer error! Wrong indices: currentPos=%d, currentOffset=%d, sentence=%s", currentPosition, bytes[currentPosition], *sent.Text)
					return errors.New(errText)
				}
				txt = (*sent.Text)[bytes[currentPosition]:]
			} else {
				endOffset := bytes[endOffsetIdx]
				txt = (*sent.Text)[bytes[currentPosition]:endOffset]
			}

			begin := int32(currentPosition) + sent.Begin
			end := begin + int32(tokenLen)
			newToken := createToken(txt, begin, end, IsWord, IsNumber, IsSymbol, IsNewline, IsPunct)
			sent.Tokens = append(sent.Tokens, newToken)
		}
		currentPosition += tokenLen
		currentPosition = findFirstCharOfNextToken(currentPosition, runes)
	}

	return nil
}

func findFirstCharOfNextToken(startPosition int, runes []rune) int {

	for position := startPosition; position < len(runes); position++ {
		c := runes[position]
		if !utils.IsWhiteSpace(c) {
			return position
		}

		if c == '\n' || c == '\r' {
			return position
		}

	}

	return -1
}

func isTelephoneNumberChar(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '-'
}

func isNumericChar(ch rune) bool {
	return unicode.IsDigit(ch) || ch == ',' || ch == '.'
}

func getLengthIfIsNumberThatStartsWithPeriod(currentPosition int, runes []rune) int {
	l := len(runes) - currentPosition
	if l < 2 {
		return -1
	}

	index := currentPosition + 1
	ch := runes[index]

	if !unicode.IsDigit(ch) {
		return -1
	}

	index++
	for index < currentPosition+l {
		ch = runes[index]
		if !unicode.IsDigit(ch) {
			return index - currentPosition
		}
		index++
	}

	return l
}

const ellipsis = "..."

func isEllipsis(bytesPosition int, txt string) bool {
	return strings.HasPrefix(txt[bytesPosition:], ellipsis)
}

func lenIfIsTelephoneNumber(currentPosition int, runes []rune, nextNonTelephoneNumberChar int) int {
	if nextNonTelephoneNumberChar < 0 {
		return nextNonTelephoneNumberChar
	}

	l := nextNonTelephoneNumberChar - currentPosition

	s := runes[currentPosition:nextNonTelephoneNumberChar]
	// extension like 4-5555
	// or without area code like 555-1212
	// or with area code 507-555-1212
	// or with 1, like 1-507-555-1212
	// or like example in guidelines like 02-2348-2192

	if l == 6 {
		if !unicode.IsDigit(s[0]) {
			return -1
		}
		if s[1] != dash {
			return -1
		}
		if !unicode.IsDigit(s[2]) {
			return -1
		}
		if !unicode.IsDigit(s[3]) {
			return -1
		}
		if !unicode.IsDigit(s[4]) {
			return -1
		}
		if !unicode.IsDigit(s[5]) {
			return -1
		}
		return l
	} else if l == 8 {
		if !unicode.IsDigit(s[0]) {
			return -1
		}
		if !unicode.IsDigit(s[1]) {
			return -1
		}
		if !unicode.IsDigit(s[2]) {
			return -1
		}
		if s[3] != dash {
			return -1
		}
		if !unicode.IsDigit(s[4]) {
			return -1
		}
		if !unicode.IsDigit(s[5]) {
			return -1
		}
		if !unicode.IsDigit(s[6]) {
			return -1
		}
		if !unicode.IsDigit(s[7]) {
			return -1
		}
		return l
	} else if l == 12 { // two possible formats
		// first check  507-555-1212 format
		if !unicode.IsDigit(s[0]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[1]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[2]) {
			return checkFormat2(s)
		}
		if s[3] != dash {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[4]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[5]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[6]) {
			return checkFormat2(s)
		}
		if s[7] != dash {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[8]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[9]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[10]) {
			return checkFormat2(s)
		}
		if !unicode.IsDigit(s[11]) {
			return checkFormat2(s)
		}
		return l
	} else if l == 14 { // 1-507-555-1212
		if !unicode.IsDigit(s[0]) {
			return -1
		}
		if s[1] != dash {
			return -1
		}
		if !unicode.IsDigit(s[2]) {
			return -1
		}
		if !unicode.IsDigit(s[3]) {
			return -1
		}
		if !unicode.IsDigit(s[4]) {
			return -1
		}
		if s[5] != dash {
			return -1
		}
		if !unicode.IsDigit(s[6]) {
			return -1
		}
		if !unicode.IsDigit(s[7]) {
			return -1
		}
		if !unicode.IsDigit(s[8]) {
			return -1
		}
		if s[9] != dash {
			return -1
		}
		if !unicode.IsDigit(s[10]) {
			return -1
		}
		if !unicode.IsDigit(s[11]) {
			return -1
		}
		if !unicode.IsDigit(s[12]) {
			return -1
		}
		if !unicode.IsDigit(s[13]) {
			return -1
		}
		return l
	} else {
		return -1
	}
}

func checkFormat2(s []rune) int { // 02-2348-2192
	if !unicode.IsDigit(s[0]) {
		return -1
	}
	if !unicode.IsDigit(s[1]) {
		return -1
	}
	if s[2] != dash {
		return -1
	}
	if !unicode.IsDigit(s[3]) {
		return -1
	}
	if !unicode.IsDigit(s[4]) {
		return -1
	}
	if !unicode.IsDigit(s[5]) {
		return -1
	}
	if !unicode.IsDigit(s[6]) {
		return -1
	}
	if s[7] != dash {
		return -1
	}
	if !unicode.IsDigit(s[8]) {
		return -1
	}
	if !unicode.IsDigit(s[9]) {
		return -1
	}
	if !unicode.IsDigit(s[10]) {
		return -1
	}
	if !unicode.IsDigit(s[11]) {
		return -1
	}

	return -1
}

func lenIfIsPostalCode(currentPosition int, runes []rune, nextNonPostalCodeChar int) int {
	if nextNonPostalCodeChar < 0 {
		return nextNonPostalCodeChar
	}

	l := nextNonPostalCodeChar - currentPosition

	s := runes[currentPosition:nextNonPostalCodeChar]
	// 55901-0000

	if l == 10 { // 55901-0001
		if !unicode.IsDigit(s[0]) {
			return -1
		}
		if !unicode.IsDigit(s[1]) {
			return -1
		}
		if !unicode.IsDigit(s[2]) {
			return -1
		}
		if !unicode.IsDigit(s[3]) {
			return -1
		}
		if !unicode.IsDigit(s[4]) {
			return -1
		}
		if s[5] != dash {
			return -1
		}
		if !unicode.IsDigit(s[6]) {
			return -1
		}
		if !unicode.IsDigit(s[7]) {
			return -1
		}
		if !unicode.IsDigit(s[8]) {
			return -1
		}
		if !unicode.IsDigit(s[9]) {
			return -1
		}
		return l
	} else {
		return -1
	}
}

func lenIfIsUrl(currentPosition int, runes []rune, endOfInputToConsider int) int {

	urlStarters := []string{"http://", "https://", "ftp://", "mailto:"}

	potentialUrl := string(runes[currentPosition:endOfInputToConsider])
	for _, s := range urlStarters {
		if strings.HasPrefix(potentialUrl, s) && len(potentialUrl) > len(s) {
			return endOfInputToConsider - currentPosition
		}
	}

	return -1
}

const validOtherEmailAddressCharacters = "!#$%&'*+/=?^_`{|}~-"

func lenIfIsEmailAddress(currentPosition int, runes []rune, endOfInputToConsider int) int {

	maxLenLocalPart := 64
	maxTotalLen := 320

	indexOfAt := utils.IndexOfRune(runes[currentPosition:endOfInputToConsider], at, 0)
	if indexOfAt < 1 || currentPosition+indexOfAt+1 == endOfInputToConsider || indexOfAt > maxLenLocalPart {
		return -1
	}

	for i := currentPosition; i < currentPosition+indexOfAt; i++ {
		ch := runes[i]

		if !(unicode.IsLetter(ch) || unicode.IsDigit(ch)) && !strings.ContainsRune(validOtherEmailAddressCharacters, ch) {
			return -1
		}
		if ch == period && (i == currentPosition || i == currentPosition+indexOfAt-1) {
			return -1
		}
	}

	prev := at

	for i := currentPosition + indexOfAt + 1; i < endOfInputToConsider; i++ {
		ch := runes[i]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {

		} else if ch == hyphenOrMinusSign || ch == period {

			if i+1 < endOfInputToConsider && (unicode.IsLetter(runes[i+1]) || unicode.IsDigit(runes[i+1])) {

			} else if unicode.IsLetter(prev) || unicode.IsDigit(prev) {
				return i - currentPosition - 1
			} else {
				return -1
			}
		} else {
			if unicode.IsLetter(prev) || unicode.IsDigit(prev) {
				return i - currentPosition - 1
			} else {
				return -1
			}
		}
	}

	l := endOfInputToConsider - currentPosition

	if l > maxTotalLen {
		return -1
	}

	return l
}

const www = "www."

func lenIfIsAbbreviation(currentPosition int, runes []rune, afterEndOfInputToConsider int) int {

	containsLetter := false

	if afterEndOfInputToConsider-currentPosition >= 4 && string(runes[currentPosition:currentPosition+4]) == www {
		return -1
	}
	for i := currentPosition; i < afterEndOfInputToConsider; i++ {
		ch := runes[i]
		var peekAhead rune
		if i+1 < afterEndOfInputToConsider {
			peekAhead = runes[i+1]
		} else {
			peekAhead = ' '
		}

		if unicode.IsLetter(ch) {
			containsLetter = true
		} else if ch != period {
			return -1
		} else if !containsLetter || (i+1 == len(runes)) {
			return -1
		} else {

			soFar := i + 1 - currentPosition
			l := lenIfIsAbbreviation(i+1, runes, afterEndOfInputToConsider)

			if l > 0 {
				return soFar + l
			}

			if utils.IsWhiteSpace(peekAhead) || isPossibleFinalPunctuation(peekAhead) {
				return soFar
			} else if !(unicode.IsLetter(peekAhead) || unicode.IsDigit(peekAhead)) {
				return soFar - 1
			}

			return -1

		}
	}

	return -1

}

const possibleFinalPunctuation = "?!:"

func isPossibleFinalPunctuation(c rune) bool {
	return strings.ContainsRune(possibleFinalPunctuation, c)
}

// returns (IsNumber, IsWord) values
func wordTokenOrNumToken(runes []rune, currentPosition int, tokenLen int) (bool, bool) {
	IsNumber := true
	IsWord := false

	for _, ch := range runes[currentPosition : currentPosition+tokenLen] {
		if unicode.IsLetter(ch) {
			IsNumber = false
			IsWord = true
			break
		}
	}

	return IsNumber, IsWord
}

func lenIfIsNumberContainingComma(currentPosition int, runes []rune, nextNonNumericChar int) int {
	s := runes[:nextNonNumericChar]

	commaPosition := utils.IndexOfRune(s, comma, currentPosition)
	if commaPosition < 0 {
		return -1
	}
	if commaPosition > nextNonNumericChar {
		return -1
	}

	l := -1

	periodPosition := utils.IndexOfRune(s, period, currentPosition)
	endOfWholeNumberPart := periodPosition
	if endOfWholeNumberPart < 0 {
		endOfWholeNumberPart = len(s)
	}

	if commaPosition > endOfWholeNumberPart {
		return -1
	}

	if commaPosition == 0 {
		return -1
	}

	position := commaPosition

	didNotFindExactlyThreeDigitsAfterComma := false

	for !didNotFindExactlyThreeDigitsAfterComma {
		l = position - currentPosition
		if position < endOfWholeNumberPart && s[position] == comma {
			position++
		}

		for i := 0; i < 3; i++ {
			if position < endOfWholeNumberPart && unicode.IsDigit(s[position]) {
				position++
			} else {
				didNotFindExactlyThreeDigitsAfterComma = true
			}
		}
		if position < endOfWholeNumberPart && unicode.IsDigit(s[position]) {
			didNotFindExactlyThreeDigitsAfterComma = true
		}
	}

	if l <= 0 {
		return -1
	}

	if periodPosition != len(runes)-1 && periodPosition == currentPosition+l {
		l++
		for l < nextNonNumericChar-currentPosition && unicode.IsDigit(s[currentPosition+l]) {
			l++
		}
	}

	return l
}

func getLenToNextNonDigit(s []rune, startingPosition int) int {
	for i := 0; startingPosition+i < len(s); i++ {
		ch := s[startingPosition+i]
		if !unicode.IsDigit(ch) {
			return i
		}
	}
	return len(s) - startingPosition
}
