package pipeline

import (
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/tokenizer"
	"text2phenotype.com/fdl/types"
	"sync"
)

type Tokenizer func(in <-chan types.Sentence) <-chan types.Sentence

func NewTokenizer() (Tokenizer, error) {
	tokenizerPTB := tokenizer.NewTokenizerPTB()
	fdlLogger := logger.NewLogger("Tokenizer PTB")

	return func(in <-chan types.Sentence) <-chan types.Sentence {
		out := make(chan types.Sentence)

		go func() {
			defer close(out)
			var wg sync.WaitGroup
			for sent := range in {
				wg.Add(1)
				go func(sent types.Sentence) {

					defer wg.Done()
					err := tokenizerPTB(&sent)
					if err != nil {
						fdlLogger.Error().Err(err)
					}

					out <- sent
				}(sent)

			}

			wg.Wait()
		}()

		return out
	}, nil
}
