package pipeline

import (
	"text2phenotype.com/fdl/lookup"
	"text2phenotype.com/fdl/smoking"
	"text2phenotype.com/fdl/types"
	"strings"
	"unicode/utf8"
)

type Result struct {
	ConfigName string
	Data       interface{}
}

func NewDefaultClinicalResult() func(in <-chan []types.Annotation, cfg LookupConfig, request Request) <-chan Result {
	ttyIdx := types.TTY

	return func(in <-chan []types.Annotation, cfg LookupConfig, request Request) <-chan Result {
		out := make(chan Result)
		go func() {
			defer close(out)
			var allAnnotation []types.Annotation

			for annotations := range in {
				allAnnotation = append(allAnnotation, annotations...)
			}

			var response types.DefaultClinicalResponse
			response.Content = make([]types.ContentSection, len(allAnnotation))
			response.DocId = request.Tid

			offsetEnd := int32(utf8.RuneCountInString(request.Text))

			// make concepts
			for i, ann := range allAnnotation {

				// make umls concepts
				umlsConcepts := make([]types.UmlsConcept, len(ann.Concepts))
				for j, concept := range ann.Concepts {

					// convert tui pointers to strings
					tuis := make([]string, len(concept.TUI))
					for tIdx, tui := range concept.TUI {
						tuis[tIdx] = strings.ToUpper(tui)
					}

					// create sab concepts
					sabConcepts := make([]types.SabConcept, len(concept.Codes))
					scIdx := 0
					for codingScheme, codes := range concept.Codes {

						vocabConcepts := make([]types.VocabConcept, len(codes))
						vcIdx := 0
						for code, params := range codes {
							tty, hasTty := params[ttyIdx]
							if hasTty {
								strTty := make([]string, len(tty))
								for ttyIdx, ttyVal := range tty {
									strTty[ttyIdx] = strings.ToUpper(ttyVal)
								}
								vocabConcepts[vcIdx].Tty = strTty
							} else {
								vocabConcepts[vcIdx].Tty = []string{}
							}

							vocabConcepts[vcIdx].Code = code

							vcIdx++
						}

						sabConcepts[scIdx] = types.SabConcept{
							CodingScheme:  strings.ToUpper(codingScheme),
							VocabConcepts: vocabConcepts,
						}

						scIdx++
					}

					umlsConcepts[j] = types.UmlsConcept{
						Tui:           tuis,
						Cui:           strings.ToUpper(*concept.CUI),
						PreferredText: concept.PrefText,
						SabConcepts:   sabConcepts,
					}
				}

				// make content section
				contentItem := types.ContentSection{
					Id: i,
					Sentence: []int32{
						ann.Sentence.Begin,
						ann.Sentence.End,
					},
					Text: []interface{}{
						*ann.Text,
						ann.Begin,
						ann.End,
					},
					SectionOid:    "SIMPLE_SEGMENT",
					Attributes:    ann.Attributes,
					Aspect:        lookup.GetAspect(ann.Semantic),
					Name:          ann.GetName(),
					UmlsConcepts:  umlsConcepts,
					SectionOffset: []int32{0, offsetEnd},
				}

				response.Content[i] = contentItem

			}

			out <- Result{
				ConfigName: cfg.Name,
				Data:       response,
			}
		}()
		return out
	}
}

func NewSmokingStatusResult() func(in <-chan types.Sentence, key string, request Request) <-chan Result {
	return func(in <-chan types.Sentence, key string, request Request) <-chan Result {
		out := make(chan Result)

		go func() {

			defer close(out)
			var smokingResolver documentSmokingStatusResolver

			var response types.SmokingStatusResponse
			for sent := range in {
				response.DocId = request.Tid
				smokingResolver.AddStatus(sent.Attributes.SmokingStatus)
				response.Sentences = append(response.Sentences, types.SmokingStatusSection{
					Status: sent.Attributes.SmokingStatus,
					Text: []interface{}{
						*sent.Text,
						sent.Begin,
						sent.End,
					},
				})
			}

			response.SmokingStatus = smokingResolver.Resolve()

			out <- Result{
				ConfigName: key,
				Data:       response,
			}
		}()

		return out
	}
}

type documentSmokingStatusResolver struct {
	unknownCnt   int
	currentCnt   int
	pastCnt      int
	smokerCnt    int
	nonSmokerCnt int
}

func (resolver *documentSmokingStatusResolver) AddStatus(status string) {
	switch status {
	case smoking.ClassCurrSmoker:
		resolver.currentCnt++
	case smoking.ClassNonSmoker:
		resolver.nonSmokerCnt++
	case smoking.ClassPastSmoker:
		resolver.pastCnt++
	case smoking.ClassSmoker:
		resolver.smokerCnt++
	default:
		resolver.unknownCnt++
	}
}
func (resolver documentSmokingStatusResolver) Resolve() string {
	switch {
	case resolver.unknownCnt > 0 && resolver.smokerCnt == 0 && resolver.pastCnt == 0 && resolver.currentCnt == 0 && resolver.nonSmokerCnt == 0:
		return smoking.ClassUnknown
	case resolver.nonSmokerCnt >= 1 && resolver.unknownCnt >= 0 && resolver.pastCnt == 0 && resolver.currentCnt == 0 && resolver.smokerCnt == 0:
		return smoking.ClassNonSmoker
	case resolver.currentCnt >= 1:
		return smoking.ClassCurrSmoker
	case resolver.pastCnt >= 1 && resolver.currentCnt <= 0:
		return smoking.ClassPastSmoker
	case resolver.smokerCnt >= 1 && resolver.currentCnt <= 0 && resolver.pastCnt <= 0:
		return smoking.ClassSmoker
	}

	return smoking.ClassUnknown
}
