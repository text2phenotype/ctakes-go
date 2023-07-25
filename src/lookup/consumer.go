package lookup

import (
	"text2phenotype.com/fdl/types"
)

type Consumer func(
	spans []types.Span,
	cuis [][]*string,
	conceptMap map[*string]SemanticConcepts,
) []types.Annotation

func GetUsedSemantics(cuiConcepts map[*string]SemanticConcepts) []types.Semantic {
	usedSemanticTypes := make(map[types.Semantic]bool)
	for _, semanticConcepts := range cuiConcepts {
		for semanticType := range semanticConcepts {
			usedSemanticTypes[semanticType] = true
		}
	}

	result := make([]types.Semantic, len(usedSemanticTypes))
	i := 0
	for semanticType := range usedSemanticTypes {
		result[i] = semanticType
		i++
	}
	return result
}

func HasSemantic(semantic types.Semantic, concepts []types.Concept) bool {
	for _, concept := range concepts {
		for _, sem := range concept.Semantics {
			if sem == semantic {
				return true
			}
		}
	}
	return false
}

func CombineSlices(sliceA []int, sliceB []int) []int {
	hashBuf := make(map[int]bool)

	for _, n := range sliceA {
		hashBuf[n] = true
	}

	for _, n := range sliceB {
		hashBuf[n] = true
	}

	resultList := make([]int, len(hashBuf))
	i := 0
	for n := range hashBuf {
		resultList[i] = n
		i++
	}
	return resultList
}

func CreatePresideTerms(textSpans []types.Span, cuis [][]*string) ([]types.Span, [][]*string) {

	discardSpans := make(map[types.Span]bool)

	count := len(textSpans)
	for i := 0; i < count; i++ {
		spanKeyI := textSpans[i]
		for j := i + 1; j < count; j++ {
			spanKeyJ := textSpans[j]
			if (spanKeyJ.Begin <= spanKeyI.Begin && spanKeyJ.End > spanKeyI.End) || (spanKeyJ.Begin < spanKeyI.Begin && spanKeyJ.End >= spanKeyI.End) {
				discardSpans[spanKeyI] = true
				break
			}

			if (spanKeyI.Begin <= spanKeyJ.Begin && spanKeyI.End > spanKeyJ.End) || (spanKeyI.Begin < spanKeyJ.Begin && spanKeyI.End >= spanKeyJ.End) {
				discardSpans[spanKeyJ] = true
			}
		}
	}

	presideSpans := make([]types.Span, count-len(discardSpans))
	presideCuis := make([][]*string, count-len(discardSpans))

	presideIdx := 0
	for spanIdx, span := range textSpans {
		if isOk := discardSpans[span]; !isOk {
			presideSpans[presideIdx] = span
			presideCuis[presideIdx] = cuis[spanIdx]
			presideIdx++
		}
	}

	return presideSpans, presideCuis
}

func CreateConsumer(precisionMode bool) Consumer {

	consumeTypeIdHits := func(
		semantic types.Semantic,
		spans []types.Span,
		cuis [][]*string,
		conceptMap map[*string]SemanticConcepts,
		annotations *[]types.Annotation) {

		if precisionMode {
			spans, cuis = CreatePresideTerms(spans, cuis)
		}

		for i, span := range spans {
			spanCuis := cuis[i]

			var concepts []*types.Concept

			for _, cui := range spanCuis {

				semantics, hasCUI := conceptMap[cui]
				if hasCUI {
					semanticConcepts, hasSemantic := semantics[semantic]
					if hasSemantic {
						concepts = append(concepts, semanticConcepts)
					}
				}
			}

			ann := types.Annotation{
				Semantic:   semantic,
				Concepts:   concepts,
				Attributes: make(map[string]interface{}),
			}
			ann.Begin = span.Begin
			ann.End = span.End
			ann.Text = span.Text
			*annotations = append(*annotations, ann)
		}

	}

	return func(spans []types.Span, spanCuis [][]*string, conceptMap map[*string]SemanticConcepts) []types.Annotation {
		semantics := GetUsedSemantics(conceptMap)

		var annotation = make([]types.Annotation, 0)
		for _, semantic := range semantics {

			var semanticSpans []types.Span
			var semanticCuis [][]*string

			for i, span := range spans {
				cuis := spanCuis[i]

				semanticSpanCuis := make(map[*string]bool)
				for _, cui := range cuis {
					semanticConcepts, ok := conceptMap[cui]
					if !ok {
						continue
					}
					_, hasSemantic := semanticConcepts[semantic]
					if hasSemantic {
						semanticSpanCuis[cui] = true
					}
				}

				if len(semanticSpanCuis) > 0 {
					semanticSpans = append(semanticSpans, span)

					cuiSet := make([]*string, len(semanticSpanCuis))
					i := 0
					for cui := range semanticSpanCuis {
						cuiSet[i] = cui
						i++
					}
					semanticCuis = append(semanticCuis, cuiSet)
				}
			}
			consumeTypeIdHits(semantic, semanticSpans, semanticCuis, conceptMap, &annotation)
		}

		return annotation
	}
}
