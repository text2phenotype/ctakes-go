package drug_ner

import (
	"text2phenotype.com/fdl/drug_ner/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"errors"
	"sort"
)

const (
	paramMedFrequencyNumber = "medFrequencyNumber"
	paramMedFrequencyUnit   = "medFrequencyUnit"
	paramMedStrengthNum     = "medStrengthNum"
	paramMedStrengthUnit    = "medStrengthUnit"
	paramMedStatusChange    = "medStatusChange"
	paramMedDosage          = "medDosage"
	paramMedRoute           = "medRoute"
	paramMedForm            = "medForm"
	paramMedDuration        = "medDuration"
)

type fsmResult struct {
	fractions      types.Spans
	ranges         types.Spans
	dosages        types.Spans
	suffixes       types.Spans
	durations      types.Spans
	routes         types.Spans
	frequencies    types.Spans
	statuses       types.Spans
	decimals       types.Spans
	strength       types.Spans
	strengthUnits  types.Spans
	frequencyUnits types.Spans
	forms          types.Spans
}

func (res *fsmResult) sort() {
	sort.Sort(res.fractions)
	sort.Sort(res.ranges)
	sort.Sort(res.dosages)
	sort.Sort(res.suffixes)
	sort.Sort(res.durations)
	sort.Sort(res.routes)
	sort.Sort(res.frequencies)
	sort.Sort(res.statuses)
	sort.Sort(res.decimals)
	sort.Sort(res.strength)
	sort.Sort(res.strengthUnits)
	sort.Sort(res.frequencyUnits)
	sort.Sort(res.forms)

}

func NewDrugNerAttributesExtractor(maxAttributeDistance int, params fsm.Params) (DrugNerAttributesExtractor, error) {
	result := drugNerAttributesExtractor{
		fractionFSM:          fsm.NewFractionStrengthFSM(params.FractionFSM),
		rangeFSM:             fsm.NewRangeStrengthFSM(params.RangeFSM),
		subMedSectionFSM:     fsm.NewSubSectionIndicatorFSM(params.SubMedSectionFSM),
		dosagesFSM:           fsm.NewDosagesFSM(params.DosageFSM),
		suffixFSM:            fsm.NewSuffixStrengthFSM(params.SuffixFSM),
		durationFSM:          fsm.NewDurationFSM(params.DurationFSM),
		routeFSM:             fsm.NewRouteFSM(params.RouteFSM),
		frequencyFSM:         fsm.NewFrequencyFSM(params.FrequencyFSM),
		statusFSM:            fsm.NewDrugChangeStatusFSM(params.ChangeStatusFSM),
		decimalFSM:           fsm.NewDecimalStrengthFSM(),
		strengthFSM:          fsm.NewStrengthFSM(params.StrengthFSM),
		strengthUnitFSM:      fsm.NewStrengthUnitFSM(params.StrengthUnitFSM),
		frequencyUnitFSM:     fsm.NewFrequencyUnitFSM(params.FrequencyUnitFSM),
		formFSM:              fsm.NewFormFSM(params.FormFSM),
		timeFSM:              fsm.NewTimeFSM(params.TimeFSM),
		maxAttributeDistance: 10,
	}
	if maxAttributeDistance > 0 {
		result.maxAttributeDistance = maxAttributeDistance
	}
	return &result, nil
}

type DrugNerAttributesExtractor interface {
	Extract(annotations []*types.Annotation) error
}

type drugNerAttributesExtractor struct {
	fractionFSM      fsm.FractionStrengthFSM
	rangeFSM         fsm.RangeStrengthFSM
	subMedSectionFSM fsm.SubSectionIndicatorFSM
	dosagesFSM       fsm.DosagesFSM
	suffixFSM        fsm.SuffixStrengthFSM
	durationFSM      fsm.DurationFSM
	routeFSM         fsm.RouteFSM
	frequencyFSM     fsm.FrequencyFSM
	statusFSM        fsm.DrugChangeStatusFSM
	decimalFSM       fsm.DecimalStrengthFSM
	strengthFSM      fsm.StrengthFSM
	strengthUnitFSM  fsm.StrengthUnitFSM
	frequencyUnitFSM fsm.FrequencyUnitFSM
	formFSM          fsm.FormFSM
	timeFSM          fsm.TimeFSM

	maxAttributeDistance int
}

func (extractor *drugNerAttributesExtractor) Extract(annotations []*types.Annotation) error {
	sentMap := make(map[*types.Sentence][]*types.Annotation)
	for _, ann := range annotations {
		if ann.Semantic != types.SemanticDrug {
			continue
		}

		sent := ann.Sentence
		if sent == nil {
			errText := "sentence is nil"
			return errors.New(errText)
		}

		annList := sentMap[sent]
		annList = append(annList, ann)
		sentMap[sent] = annList
	}

	for sent, drugMentions := range sentMap {
		if len(drugMentions) == 0 {
			continue
		}
		extractor.searchDrugAttributes(sent, drugMentions)
	}

	return nil
}

func (extractor *drugNerAttributesExtractor) executeFSMs(sent *types.Sentence) *fsmResult {
	var result fsmResult

	timeTokens := extractor.timeFSM(sent)
	result.fractions = extractor.fractionFSM(sent, []types.HasSpan{})
	result.decimals = extractor.decimalFSM(sent)
	result.statuses = extractor.statusFSM(sent)
	result.ranges = extractor.rangeFSM(sent)
	result.strengthUnits = extractor.strengthUnitFSM(sent, result.ranges)
	result.forms = extractor.formFSM(sent, []types.HasSpan{})
	result.strength = extractor.strengthFSM(sent, result.strengthUnits, result.fractions)
	result.dosages = extractor.dosagesFSM(sent, result.forms, result.strengthUnits)
	result.suffixes = extractor.suffixFSM(sent, result.strengthUnits)
	result.routes = extractor.routeFSM(sent)
	result.frequencyUnits = extractor.frequencyUnitFSM(sent, timeTokens)
	result.frequencies = extractor.frequencyFSM(sent, result.frequencyUnits, result.ranges)
	result.durations = extractor.durationFSM(sent, result.ranges)

	result.sort()
	return &result
}

func (extractor *drugNerAttributesExtractor) searchDrugAttributes(sent *types.Sentence, drugMentions []*types.Annotation) {
	fsmRes := extractor.executeFSMs(sent)
	spans := extractor.getWindowSpans(sent, drugMentions)

	runes, _ := utils.MakeRuneByteSlices(*sent.Text)

	for i, span := range spans {
		drug := drugMentions[i]

		freqUnitTokIdx, isFreqUnitTokOk := fsmRes.frequencyUnits.SearchFirstInSpan(span[0], span[1])
		if isFreqUnitTokOk {
			tok := fsmRes.frequencyUnits[freqUnitTokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedFrequencyUnit] = []interface{}{annTxt, tok.Begin, tok.End}
		} else {
			drug.Attributes[paramMedFrequencyUnit] = []interface{}{}
		}

		tokIdx, isOk := fsmRes.frequencies.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.frequencies[tokIdx].(*fsm.FrequencyToken)
			annTxt := tok.Value
			if len(annTxt) == 0 {
				annTxt = string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			}
			drug.Attributes[paramMedFrequencyNumber] = []interface{}{annTxt, tok.Begin, tok.End}
		} else {
			if !isFreqUnitTokOk {
				drug.Attributes[paramMedFrequencyNumber] = []interface{}{}
			} else {
				freqUnitTok := fsmRes.frequencyUnits[freqUnitTokIdx].(*fsm.FrequencyUnitToken)
				if freqUnitTok.Quantity != fsm.QuantityPrn {
					drug.Attributes[paramMedFrequencyNumber] = []interface{}{freqUnitTok.Quantity, freqUnitTok.Begin, freqUnitTok.End}
				}
			}
		}

		tokIdx, isOk = fsmRes.strength.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.strength[tokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedStrengthNum] = []interface{}{annTxt, tok.Begin, tok.End}
		} else {
			drug.Attributes[paramMedStrengthNum] = []interface{}{}
		}

		tokIdx, isOk = fsmRes.strengthUnits.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.strengthUnits[tokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedStrengthUnit] = []interface{}{annTxt, tok.Begin, tok.End}
		} else {
			drug.Attributes[paramMedStrengthUnit] = []interface{}{}
		}

		tokIdx, isOk = fsmRes.statuses.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.statuses[tokIdx].(*fsm.DrugChangeStatusToken)
			drug.Attributes[paramMedStatusChange] = tok.Status
		} else {
			drug.Attributes[paramMedStatusChange] = nil
		}

		tokIdx, isOk = fsmRes.dosages.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.dosages[tokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedDosage] = annTxt
		} else {
			drug.Attributes[paramMedDosage] = nil
		}

		tokIdx, isOk = fsmRes.routes.SearchFirstInSpan(span[0], span[1])
		if isOk {
			//tok := fsmRes.routes[tokIdx].GetSpan()
			//annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			//drug.Attributes[paramMedRoute] = []interface{}{annTxt, tok.Begin, tok.End}
			tok := fsmRes.routes[tokIdx].(*fsm.RouteToken)
			drug.Attributes[paramMedRoute] = tok.FormMethod
		} else {
			drug.Attributes[paramMedRoute] = nil
		}

		tokIdx, isOk = fsmRes.forms.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.forms[tokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedForm] = annTxt
		} else {
			drug.Attributes[paramMedForm] = nil
		}

		tokIdx, isOk = fsmRes.durations.SearchFirstInSpan(span[0], span[1])
		if isOk {
			tok := fsmRes.durations[tokIdx].GetSpan()
			annTxt := string(runes[tok.Begin-sent.Begin : tok.End-sent.Begin])
			drug.Attributes[paramMedDuration] = annTxt
		} else {
			drug.Attributes[paramMedDuration] = nil
		}
	}

}

func (extractor *drugNerAttributesExtractor) getWindowSpans(sent *types.Sentence, drugMentions []*types.Annotation) [][2]int32 {
	spans := make([][2]int32, len(drugMentions))
	tokens := sent.Tokens

	for i, drug := range drugMentions {
		span := [2]int32{drug.End, sent.End}
		if i < len(drugMentions)-1 {
			span[1] = drugMentions[i+1].Begin
		}

		startTokenIdx := 0
		// move to token after drug mention
		for startTokenIdx < len(tokens) && tokens[startTokenIdx].Begin < drug.End {
			startTokenIdx++
		}

		// save position and move to span end ot new line token
		totalTokens := startTokenIdx
		for totalTokens < len(tokens) && tokens[totalTokens].End <= span[1] && totalTokens <= startTokenIdx+extractor.maxAttributeDistance {
			totalTokens++
		}

		span[1] = tokens[totalTokens-1].End

		spans[i] = span
	}

	return spans
}
