package pipeline

import (
	"text2phenotype.com/fdl/pos"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"sync"
)

type Tagger func(tokens []*types.Token) []*types.Token

func NewPOSTagger(model pos.Model) func(in <-chan types.Sentence) <-chan types.Sentence {
	tagger := pos.NewTagger(model)

	return func(in <-chan types.Sentence) <-chan types.Sentence {
		out := make(chan types.Sentence)
		go func() {
			defer close(out)
			var wg sync.WaitGroup
			for sent := range in {

				wg.Add(1)
				go func(sent types.Sentence) {
					defer wg.Done()
					if sent.Tokens != nil && len(sent.Tokens) > 0 {

						words := make([]*types.Token, 0, len(sent.Tokens))
						wordsIndex := make([]int, 0, len(sent.Tokens))
						for i, token := range sent.Tokens {
							if token.IsNewline {
								continue
							}
							words = append(words, token)
							wordsIndex = append(wordsIndex, i)
						}

						tags := tagger(words)

						for i, tokenTag := range tags {
							tokenIndex := wordsIndex[i]
							sent.Tokens[tokenIndex].Tag = utils.GlobalStringStore().GetPointer(tokenTag)
						}
					}
					out <- sent
				}(sent)

			}

			wg.Wait()

		}()
		return out
	}
}
