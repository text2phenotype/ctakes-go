package pipeline

import (
	"text2phenotype.com/fdl/drug_ner"
	"text2phenotype.com/fdl/drug_ner/fsm"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/types"
	"sync"
)

func NewDrugAttributesAnnotator(params DrugNerParams, fsmParams fsm.Params) (func(in <-chan []types.Annotation) <-chan []types.Annotation, error) {

	fdlLogger := logger.NewLogger("DrugNer attributes annotator")

	extractor, err := drug_ner.NewDrugNerAttributesExtractor(params.MaxAttributeDistance, fsmParams)
	if err != nil {
		return nil, err
	}

	return func(in <-chan []types.Annotation) <-chan []types.Annotation {

		out := make(chan []types.Annotation)

		go func() {
			defer close(out)
			var wg sync.WaitGroup
			for annotations := range in {
				wg.Add(1)
				go func(annotations []types.Annotation) {
					defer wg.Done()
					if len(annotations) > 0 {
						drugMentions := make([]*types.Annotation, 0, len(annotations))
						for i := 0; i < len(annotations); i++ {
							ann := annotations[i]
							if ann.Semantic == types.SemanticDrug {
								drugMentions = append(drugMentions, &ann)
							}
						}

						if len(drugMentions) > 0 {
							err := extractor.Extract(drugMentions)
							if err != nil {
								fdlLogger.Err(err)
							}
						}
					}

					out <- annotations
				}(annotations)

			}

			wg.Wait()
		}()

		return out

	}, nil
}
