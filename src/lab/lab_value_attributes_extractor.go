package lab

import (
	"text2phenotype.com/fdl/lab/fsm"
	"text2phenotype.com/fdl/ml"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	maxWindowSize             = 50 // maximum tokens count which should be used for prediction
	lookupKeyAnnotationsCount = 5  // how many key annotations (terms or values) should be looked up for prediction
	cpos                      = "CC"

	LvDistanceFeatName      = "LV_DISTANCE"
	LvTermsBetweenFeatName  = "LV_TERMS_BETWEEN"
	LvValuesBetweenFeatName = "LV_VALUES_BETWEEN"
	LvIsClosestFeatName     = "LV_IS_CLOSEST"
	LvPointFeatName         = "LV_POINT"
	LvNewLineFeatName       = "LV_NEW_LINE"
	LvPunctuationFeatName   = "LV_PUNCTUATION"
	LvConjFeatName          = "LV_CONJ"

	LinkedCategory = "LINKED"

	attrLabValue     = "labValue"
	attrLabValueUnit = "labValueUnit"
)

type LabValuesAttributesExtractor interface {
	Extract(annotations []*types.Annotation) error
}

type labValuesAttributesExtractor struct {
	stringValues     []string
	maxTokenDistance int
	classifier       *ml.CRF
	unitFSM          fsm.LabUnitsFSM
	datesFSM         fsm.DateFSM
	rangesFSM        fsm.RangeFSM
	fractionFSM      fsm.FractionFSM
}

func (extractor *labValuesAttributesExtractor) Extract(annotations []*types.Annotation) error {

	sent := annotations[0].Sentence
	if sent == nil {
		errTxt := fmt.Sprintf("annotation '%s' [%d:%d]: sentence is nil", *annotations[0].Text, annotations[0].Begin, annotations[0].End)
		return errors.New(errTxt)
	}

	tokens := make([]types.Token, len(sent.Tokens))
	for tidx, tokenPtr := range sent.Tokens {
		tokens[tidx] = tokenPtr.Clone()
	}
	tokens = extractor.unitFSM.Split(tokens)

	// sort annotations and tokens
	annotationsIdx := make([]int, len(annotations))
	tokenIdx := make([]int, len(tokens))

	for i := 0; i < len(annotationsIdx) || i < len(tokenIdx); i++ {
		if i < len(annotationsIdx) {
			annotationsIdx[i] = i
		}

		if i < len(tokenIdx) {
			tokenIdx[i] = i
		}
	}

	sort.SliceStable(annotationsIdx, func(i, j int) bool {
		return annotations[annotationsIdx[i]].Begin < annotations[annotationsIdx[j]].Begin
	})

	sort.SliceStable(tokenIdx, func(i, j int) bool {
		return tokens[tokenIdx[i]].Begin < tokens[tokenIdx[j]].Begin
	})

	// find all special value words and create tokens for them
	var specialValueWords []*ValueToken

	txt := strings.ToLower(*sent.Text)
	txtLen := len(txt)
	for txtLen > 0 {
		for _, wrd := range extractor.stringValues {
			//offset := strings.LastIndex(txt, wrd)
			offset := strings.Index(txt, wrd)
			if offset != -1 {
				l := int32(utf8.RuneCountInString(wrd))
				txt = txt[:offset]

				beginIdx := sent.Begin + int32(utf8.RuneCountInString(txt))
				newToken := &types.Token{
					Span: types.Span{
						Begin: beginIdx,
						End:   beginIdx + l,
						Text:  utils.GlobalStringStore().GetPointer(wrd),
					},
				}

				newValueToken := ValueToken{newToken}

				specialValueWords = append([]*ValueToken{&newValueToken}, specialValueWords...)
			}
		}
		if txtLen == len(txt) {
			break
		}

		txtLen = len(txt)
	}

	fractions := extractor.fractionFSM(*sent)
	ranges := extractor.rangesFSM(*sent)
	dates := extractor.datesFSM(*sent, ranges)

	// built sequence of lab mentions and tokens
	valuesIndices := make([]int, 0, len(tokens))
	labMentionsIndices := make([]int, 0, len(annotations))
	allAnnotations := make([]types.HasSpan, 0, len(tokens))
	lastSpanEnd := int32(-1)
	annotationCursor := 0
	specialWordsCursor := 0
	dateCursor := 0
	rangeCursor := 0
	fractionCursor := 0

	for tIdx := 0; tIdx < len(tokens); tIdx++ {
		tok := tokens[tokenIdx[tIdx]]

		if tok.Begin < lastSpanEnd {
			continue
		}

		if annotationCursor < len(annotations) {
			ann := annotations[annotationsIdx[annotationCursor]]
			if types.CheckSpansOverlap(&tok.Span, &ann.Span) {
				if ann.Semantic == types.SemanticLab {
					labMentionsIndices = append(labMentionsIndices, len(allAnnotations))
					allAnnotations = append(allAnnotations, ann)
					lastSpanEnd = ann.End
					annotationCursor++
					continue
				}
				annotationCursor++
			}
		}

		if dateCursor < len(dates) && tok.End > dates[dateCursor].Begin {
			if types.CheckSpansOverlap(&tok.Span, &dates[dateCursor].Span) {
				if tok.End == dates[dateCursor].End {
					dateCursor++
				}
				continue
			}
		}

		if rangeCursor < len(ranges) && tok.End > ranges[rangeCursor].Begin {
			if types.CheckSpansOverlap(&tok.Span, &ranges[rangeCursor].Span) {
				if tok.End == ranges[rangeCursor].End {
					valuesIndices = append(valuesIndices, len(allAnnotations))
					allAnnotations = append(allAnnotations, &ranges[rangeCursor])
					rangeCursor++
				}
				continue
			}
		}

		if fractionCursor < len(fractions) && tok.End > fractions[fractionCursor].Begin {
			if types.CheckSpansOverlap(&tok.Span, &fractions[fractionCursor].Span) {
				if tok.End == fractions[fractionCursor].End {
					valuesIndices = append(valuesIndices, len(allAnnotations))
					allAnnotations = append(allAnnotations, &fractions[fractionCursor])
					fractionCursor++
				}
				continue
			}
		}

		if tok.IsNumber {
			valuesIndices = append(valuesIndices, len(allAnnotations))
			allAnnotations = append(allAnnotations, &tok)
			lastSpanEnd = tok.End
		} else {
			if len(specialValueWords) > 0 && specialWordsCursor < len(specialValueWords) {
				specWrd := specialValueWords[specialWordsCursor]
				specWrdSpan := specWrd.GetSpan()
				for specialWordsCursor < len(specialValueWords) && specWrdSpan.Begin <= tok.Begin {
					specWrd = specialValueWords[specialWordsCursor]
					if types.CheckSpansOverlap(&tok.Span, specWrdSpan) {
						valuesIndices = append(valuesIndices, len(allAnnotations))
						allAnnotations = append(allAnnotations, specWrd)

						lastSpanEnd = specWrdSpan.End
						specialWordsCursor++
						break
					}
					specialWordsCursor++
				}

			}
		}

		if lastSpanEnd < tok.End {
			allAnnotations = append(allAnnotations, &tok)
		}
	}

	if len(labMentionsIndices) == 0 || len(valuesIndices) == 0 {
		return nil
	}

	pairs, allFeatures := extractPairsFeatures(labMentionsIndices, valuesIndices, allAnnotations)
	if len(allFeatures) > 0 {
		predictedCategories := extractor.classifier.Predict(allFeatures)

		for i, cat := range predictedCategories {
			if strings.EqualFold(LinkedCategory, cat) {
				labMentionIdx := pairs[i][0]
				valueIdx := pairs[i][1]
				if utils.AbsInt(labMentionIdx-valueIdx) < extractor.maxTokenDistance {
					labMention := allAnnotations[labMentionIdx].(*types.Annotation)
					value := allAnnotations[valueIdx].GetSpan()
					if labMention.Attributes == nil {
						labMention.Attributes = make(map[string]interface{})
					}
					labMention.Attributes[attrLabValue] = []interface{}{
						*value.Text,
						value.Begin,
						value.End,
					}

					if valueIdx < len(allAnnotations)-1 {
						//add unit attribute
						unit := extractor.unitFSM.Execute(allAnnotations[valueIdx+1:])
						if unit != nil {
							unitSpan := unit.GetSpan()
							labMention.Attributes[attrLabValueUnit] = []interface{}{
								*unitSpan.Text,
								unitSpan.Begin,
								unitSpan.End,
							}
						}
					}

				}
			}

		}
	}

	return nil
}

func extractPairsFeatures(labMentionsIndices []int, valuesIndices []int, allAnnotations []types.HasSpan) ([][]int, [][]ml.Feature) {
	var pairs [][]int
	var allFeatures [][]ml.Feature
	for _, aIdx := range labMentionsIndices {
		valuesWereUsed := 0
		batchPairs := make([][]int, 0, len(valuesIndices))
		batchFeatures := make([][]ml.Feature, 0, len(valuesIndices))

		for _, vIdx := range valuesIndices {
			if utils.AbsInt(aIdx-vIdx) <= maxWindowSize {

				feats := extractPairFeatures(aIdx, vIdx, allAnnotations)
				if len(feats) == 0 {
					continue
				}

				batchPairs = append(batchPairs, []int{aIdx, vIdx})
				batchFeatures = append(batchFeatures, feats)

				valuesWereUsed++
			}
			if valuesWereUsed > 2*lookupKeyAnnotationsCount {
				break
			}
		}

		for n := len(batchPairs) - 1; n >= 0; n-- {
			pairs = append(pairs, batchPairs[n])
			allFeatures = append(allFeatures, batchFeatures[n])
		}

	}

	return pairs, allFeatures
}

func extractPairFeatures(labMention int, value int, allAnnotations []types.HasSpan) []ml.Feature {
	var features []ml.Feature

	//nolint TODO: Check if this is needed
	firstSpan := allAnnotations[labMention].GetSpan()
	//nolint
	secondSpan := allAnnotations[value].GetSpan()

	begin := labMention
	end := value
	order := 1
	if end < begin {
		begin, end = end, begin
		//nolint
		firstSpan, secondSpan = secondSpan, firstSpan
		order = -1
	}

	distance := 0
	termsBetween := 0
	valuesBetween := 0

	for currentIdx := begin + 1; currentIdx < end; currentIdx++ {
		current := allAnnotations[currentIdx]

		distance++
		if _, isOk := current.(*types.Annotation); isOk {
			termsBetween++
			features = append(features, &ml.StrFeature{Name: LvPointFeatName, Value: fmt.Sprintf("TERM_%d", termsBetween)})
		} else if _, isOk := current.(*ValueToken); isOk {
			valuesBetween++
			features = append(features, &ml.StrFeature{Name: LvPointFeatName, Value: fmt.Sprintf("VALUE_%d", valuesBetween)})
		} else {
			tok := current.(*types.Token)

			if tok.IsNumber {
				valuesBetween++
				features = append(features, &ml.StrFeature{Name: LvPointFeatName, Value: fmt.Sprintf("VALUE_%d", valuesBetween)})
			}

			if tok.IsNewline {
				features = append(features, &ml.IntFeature{Name: LvNewLineFeatName, Value: distance})
				continue
			}

			if tok.IsPunct {
				features = append(features, &ml.StrFeature{Name: LvPunctuationFeatName, Value: *current.GetSpan().Text})
			}

			pos := tok.Tag
			if pos == nil {
				continue
			}

			if strings.EqualFold(*pos, cpos) {
				features = append(features, &ml.StrFeature{Name: LvConjFeatName, Value: strings.ToUpper(*current.GetSpan().Text)})
			}
		}
	}

	features = append(
		features,
		&ml.IntFeature{Name: LvDistanceFeatName, Value: distance * order},
		&ml.IntFeature{Name: LvTermsBetweenFeatName, Value: termsBetween},
		&ml.IntFeature{Name: LvValuesBetweenFeatName, Value: valuesBetween},
		&ml.BoolFeature{Name: LvIsClosestFeatName, Value: termsBetween+valuesBetween == 0},
	)
	return features
}

func NewLabValuesRelationExtractor(model string, labUnitFile string, maxTokenDistance int, stringValues []string) (LabValuesAttributesExtractor, error) {
	classifier, err := ml.LoadCRFFromFile(model)
	if err != nil {
		return nil, err
	}

	labUnitsFsm, err := fsm.NewLabUnitsFSM(labUnitFile)
	if err != nil {
		return nil, err
	}

	extractor := labValuesAttributesExtractor{
		stringValues:     stringValues,
		maxTokenDistance: maxTokenDistance,
		classifier:       classifier,
		unitFSM:          labUnitsFsm,
		datesFSM:         fsm.NewDateFSM(),
		rangesFSM:        fsm.NewRangeFSM(),
		fractionFSM:      fsm.NewFractionFSM(),
	}

	return &extractor, nil

}
