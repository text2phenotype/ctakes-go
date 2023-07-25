package pipeline

import (
	"text2phenotype.com/fdl/types"
	"sync"
)

func NewSentenceChannelSplitter(n int) func(in <-chan types.Sentence) []chan types.Sentence {

	return func(in <-chan types.Sentence) []chan types.Sentence {
		outs := make([]chan types.Sentence, n)
		// init channels
		for i := 0; i < n; i++ {
			outs[i] = make(chan types.Sentence)
		}

		go func() {
			defer closeAllChannels(outs)
			var wg sync.WaitGroup

			for sent := range in {
				wg.Add(1)
				go func(sent types.Sentence) {
					defer wg.Done()
					for _, out := range outs {
						out <- sent
					}
				}(sent)

			}

			wg.Wait()
		}()
		return outs
	}
}

func closeAllChannels(outs []chan types.Sentence) {
	for _, out := range outs {
		close(out)
	}
}

func NewAnnotationsChannelSplitter(n int) (func(annChannel <-chan []*types.Annotation) []chan []*types.Annotation, error) {

	return func(annChannel <-chan []*types.Annotation) []chan []*types.Annotation {
		result := make([]chan []*types.Annotation, n)
		for i := 0; i < n; i++ {
			result[i] = make(chan []*types.Annotation)
		}

		go func() {
			for {
				sent := <-annChannel
				for _, ch := range result {
					ch <- sent
				}

				if sent == nil {
					break
				}
			}
		}()
		return result
	}, nil
}
