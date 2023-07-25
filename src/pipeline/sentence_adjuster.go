package pipeline

import (
	"text2phenotype.com/fdl/types"
	"sort"
	"strings"
	"unicode/utf8"
)

type SentenceAdjusterParams struct {
	wordsInPattern []string
}

func mergeTexts(prevText string, currText string, numSpaces int32) string {
	var sb strings.Builder
	sb.WriteString(prevText)

	var i int32
	for i = 0; i < numSpaces; i++ {
		sb.WriteRune(' ')
	}

	sb.WriteString(currText)
	return sb.String()
}

func NewSentenceAdjuster(params SentenceAdjusterParams) func(in <-chan types.Sentence) <-chan types.Sentence {

	//fdlLogger := logger.NewLogger("Sentence adjuster")

	return func(in <-chan types.Sentence) <-chan types.Sentence {

		out := make(chan types.Sentence)

		go func() {
			defer close(out)

			allSentences := make([]types.Sentence, 0)

			// collect all sentences and sort
			for sent := range in {
				allSentences = append(allSentences, sent)
			}

			sort.SliceStable(allSentences, func(i, j int) bool {
				return types.SpanSortFunction(allSentences[i].GetSpan(), allSentences[j].GetSpan())
			})

			if len(allSentences) == 0 {
				return
			}

			var prevSent *types.Sentence
			for i := 0; i < len(allSentences); i++ {
				currSent := allSentences[i]

				if prevSent == nil {
					prevSent = &currSent
					continue
				}

				prevSentTxt := *prevSent.Text
				ch, _ := utf8.DecodeLastRuneInString(prevSentTxt)
				if ch != ':' {
					out <- *prevSent
					prevSent = &currSent
					continue
				}

				currSentTxt := *currSent.Text
				isFound := false
				for _, word := range params.wordsInPattern {
					if strings.HasPrefix(currSentTxt, word) {
						prevSent.End = currSent.End
						prevSent.Tokens = append(prevSent.Tokens, currSent.Tokens...)

						mergedText := mergeTexts(prevSentTxt, currSentTxt, currSent.Begin-prevSent.End)
						prevSent.Text = &mergedText
						isFound = true
						break
					}
				}

				out <- *prevSent
				prevSent = nil
				if !isFound {
					newSent := types.Sentence{
						Span: types.Span{
							Begin: currSent.Begin,
							End:   currSent.End,
							Text:  currSent.Text,
						},
					}
					newSent.Tokens = append(newSent.Tokens, currSent.Tokens...)
					prevSent = &newSent
					out <- newSent
				}
			}

			if prevSent != nil {
				out <- *prevSent
			}
		}()

		return out

	}
}
