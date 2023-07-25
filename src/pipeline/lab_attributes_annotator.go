package pipeline

import (
	"text2phenotype.com/fdl/lab"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/types"
	"sync"
)

func NewLabAttributesAnnotator(params LabValuesParams) (func(in <-chan []types.Annotation) <-chan []types.Annotation, error) {

	fdlLogger := logger.NewLogger("Lab attributes annotator")

	extractor, err := lab.NewLabValuesRelationExtractor(params.Model, params.LabUnitsFile, params.MaxTokenDistance, params.StringValues)
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
						labMentions := make([]*types.Annotation, 0, len(annotations))
						for i, ann := range annotations {
							if ann.Semantic == types.SemanticLab {
								labMentions = append(labMentions, &annotations[i])
							}
						}

						if len(labMentions) > 0 {
							err := extractor.Extract(labMentions)
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
