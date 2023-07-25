package nlp

import (
	"bufio"
	"text2phenotype.com/fdl/nlp/features"
	"text2phenotype.com/fdl/nlp/model"
	"text2phenotype.com/fdl/types"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type SentenceDetector func(in <-chan string) <-chan types.Sentence

const (
	OutcomeB = "B"
	OutcomeI = "I"
	OutcomeO = "O"
)

func loadTokenCounts(resPath string) (map[string]float64, error) {
	tokenCounts := make(map[string]float64)
	f, err := os.Open(path.Join(resPath, "tokenCounts.txt"))
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		token := strings.TrimSpace(parts[0])
		cnt, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		tokenCounts[token] = float64(cnt)
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return tokenCounts, nil
}

func NewSentenceDetector(resPath string) (SentenceDetector, error) {

	m, err := model.Load(path.Join(resPath, "sent_detector_cache_model.json"))
	if err != nil {
		return nil, err
	}
	tokenCounts, err := loadTokenCounts(resPath)
	if err != nil {
		return nil, err
	}
	outcomes := map[byte]string{
		1: OutcomeB,
		2: OutcomeI,
		3: OutcomeO,
	}
	return func(in <-chan string) <-chan types.Sentence {
		out := make(chan types.Sentence)

		go func() {
			defer close(out)
			for text := range in {

				if len(text) == 0 {
					continue
				}
				prevOutcome := OutcomeO

				startIdx := 0
				startOffset := 0

				var nextTokenBeginOffset, nextTokenEndOffset int
				var prevTokenBeginOffset, prevTokenEndOffset int
				var prevToken string
				var nextToken string

				tokenFeatures := features.NewTokenFeaturesBuilder(m)
				charFeatures := features.NewCharFeaturesBuilder(m)

				randColonStart := false

				var prevSent *types.Sentence
				prevSentStartOffset := 0

				rInd := 0

				for bytesOffset, curChar := range text {

					if bytesOffset >= nextTokenEndOffset {
						prevTokenBeginOffset, prevTokenEndOffset = nextTokenBeginOffset, nextTokenEndOffset
						if prevTokenEndOffset == 0 {
							prevTokenEndOffset = 1
						}
						nextTokenBeginOffset, nextTokenEndOffset = getNextToken(text, bytesOffset)

						prevToken = text[prevTokenBeginOffset:prevTokenEndOffset]
						nextToken = text[nextTokenBeginOffset:nextTokenEndOffset]

						tokenFeatures.Cleanup()
						tokenFeatures.AppendTokenFeatures(prevToken, nextToken)

						cnt := tokenCounts[strings.ToLower(nextToken)]
						rightLower := 0
						if cnt != 0 {
							rightLower = int(math.Round(math.Log(cnt)))
						}
						tokenFeatures.AppendFeature(
							features.TOKEN,
							features.RIGHT_LOWER,
							strconv.Itoa(rightLower),
							features.SUFFIX_TRUE,
						)
						prevDotless := prevToken
						if strings.HasSuffix(prevToken, ".") {
							prevDotless = prevToken[:len(prevToken)-1]
						}
						cnt = tokenCounts[prevDotless]
						leftDotless := 0
						if cnt != 0 {
							leftDotless = int(math.Round(math.Log(cnt)))
						}
						tokenFeatures.AppendFeature(
							features.TOKEN,
							features.LEFT_DOTLESS,
							strconv.Itoa(leftDotless),
							features.SUFFIX_TRUE,
						)
					}
					if prevOutcome != OutcomeO && (unicode.IsDigit(curChar) || unicode.IsLetter(curChar)) {
						prevOutcome = OutcomeI
						rInd++
						continue
					}
					charFeatures.Cleanup()
					charFeatures.Merge(tokenFeatures)

					charFeatures.AppendCharFeatures(curChar, features.CHARACTER)

					//search features in window [-3 ... 3]
					leftWindowOffset, rightWindowOffset := 0, utf8.RuneLen(curChar)

					curChar, _ := utf8.DecodeRuneInString(text[bytesOffset:])

					charFeatures.AppendCharFeatures(curChar, features.CHAR_OFFSET, features.CharOffsetFeatureValues[0])

					for i := 1; i <= 3; i++ {
						// left window
						if bytesOffset-leftWindowOffset > 0 {
							prevChar, prevCharSize := utf8.DecodeLastRuneInString(text[:bytesOffset-leftWindowOffset])
							charFeatures.AppendCharFeatures(prevChar, features.CHAR_OFFSET, features.CharOffsetFeatureValues[-i])
							leftWindowOffset += prevCharSize
						}
						// right window
						if bytesOffset+rightWindowOffset < len(text) {
							nextChar, nextCharSize := utf8.DecodeRuneInString(text[bytesOffset+rightWindowOffset:])
							charFeatures.AppendCharFeatures(nextChar, features.CHAR_OFFSET, features.CharOffsetFeatureValues[i])
							rightWindowOffset += nextCharSize
						}
					}
					charFeatures.AppendFeature(features.PREV_OUTCOME, prevOutcome)

					var outcome string
					outcome = outcomes[m.Predict(charFeatures)]

					isRandomColon := rInd > 0 && text[rInd] == ':' && text[rInd-1] == '\n'
					if isRandomColon {
						outcome = OutcomeO
						randColonStart = true
					}

					if outcome == OutcomeB {
						if !randColonStart {
							startIdx = rInd
							startOffset = bytesOffset
						}
					} else if outcome == OutcomeO && (prevOutcome == OutcomeI || prevOutcome == OutcomeB) {

						endInd := rInd
						endOffset := bytesOffset
						for endInd > startIdx && unicode.IsSpace(rune(text[endOffset-1])) {
							endInd--
							endOffset--
						}
						if endInd > startIdx {
							for startIdx < len(text) && unicode.IsSpace(rune(text[startOffset])) {
								startIdx++
								startOffset++
							}
							for endInd > 0 && unicode.IsSpace(rune(text[endOffset-1])) {
								endInd--
								endOffset++
							}
							if startIdx < endInd {
								sentTxt := text[startOffset:endOffset]
								newSent := types.Sentence{
									Span: types.Span{
										Begin: int32(startIdx),
										End:   int32(endInd),
										Text:  &sentTxt,
									},
								}
								randColonStart = false

								if prevSent != nil {
									if prevSent.End > newSent.Begin {
										newSent.Begin = prevSent.Begin

										startOffset = prevSentStartOffset
										sentTxt = text[startOffset:endOffset]

										newSent.Text = &sentTxt
									} else {
										out <- *prevSent
									}
								}
								prevSent = &newSent
								prevSentStartOffset = startOffset
							}
						}
					}

					if isRandomColon {
						startIdx = rInd
						startOffset = bytesOffset
						randColonStart = true
					}
					prevOutcome = outcome

					rInd++
				}

				if prevOutcome != OutcomeO {
					begin := startIdx
					beginOffset := startOffset
					end := utf8.RuneCountInString(text)
					endOffset := len(text)

					for begin < len(text) && unicode.IsSpace(rune(text[beginOffset])) {
						begin++
						beginOffset++
					}

					for endOffset > 0 {
						r, rLen := utf8.DecodeLastRuneInString(text[:endOffset])
						if !unicode.IsSpace(r) {
							break
						}
						end--
						endOffset -= rLen
					}

					sentTxt := text[beginOffset:endOffset]
					newSent := types.Sentence{
						Span: types.Span{
							Begin: int32(begin),
							End:   int32(end),
							Text:  &sentTxt,
						},
					}

					if prevSent != nil {
						if prevSent.End > newSent.Begin {
							newSent.Begin = prevSent.Begin

							sentTxt = text[prevSent.Begin:end]

							newSent.Text = &sentTxt
						} else {
							out <- *prevSent
						}
					}

					out <- newSent
				}
			}
		}()

		return out
	}, nil
}

func getNextToken(text string, offset int) (int, int) {
	if len(text) == 0 {
		return 0, 0
	}

	startOffset := offset
	// find next non whitespace char index
	for startOffset < len(text) {
		ch, chSize := utf8.DecodeRuneInString(text[startOffset:])
		if !unicode.IsSpace(ch) {
			break
		}
		startOffset += chSize
	}

	endOffset := startOffset
	for endOffset < len(text) {
		ch, chSize := utf8.DecodeRuneInString(text[endOffset:])
		if unicode.IsSpace(ch) {
			break
		}
		endOffset += chSize
	}
	return startOffset, endOffset
}
