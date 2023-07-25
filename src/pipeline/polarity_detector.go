package pipeline

import (
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/negation"
	"text2phenotype.com/fdl/types"
	"sort"
)

const PolarityParamName = "polarity"

func NewPolarityDetector(maxLeftScopeSize int, maxRightScopeSize int, scopes []types.Scope, boundaries map[string]bool) (func(in <-chan []types.Annotation) <-chan []types.Annotation, error) {

	fdlLogger := logger.NewLogger("Polarity detector")

	return func(in <-chan []types.Annotation) <-chan []types.Annotation {

		out := make(chan []types.Annotation)

		analyzer := negation.NewPolarityAnalyzer(maxLeftScopeSize, maxRightScopeSize, boundaries)
		go func() {
			defer close(out)
			allAnnotations := make([]types.Annotation, 0)

			for annotations := range in {
				allAnnotations = append(allAnnotations, annotations...)
			}

			sort.SliceStable(allAnnotations, func(i, j int) bool {
				return allAnnotations[i].Begin <= allAnnotations[j].Begin
			})

			polarities, err := analyzer(allAnnotations, scopes)
			if err != nil {
				fdlLogger.Err(err)
			} else {
				for i, ann := range allAnnotations {
					ann.Attributes[PolarityParamName] = polarities[i].Name()
				}

				out <- allAnnotations
			}
		}()

		return out

	}, nil
}
