package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type DosageFSMParams struct {
	TextNumberSet map[string]bool `drug:"text_number_set.txt"`
	SoloTextSet   map[string]bool `drug:"solo_text_set.txt"`
}

type DosageToken struct {
	types.Span
}

func (token DosageToken) GetSpan() *types.Span {
	return &token.Span
}

type DosagesFSM func(sent *types.Sentence, overrideSet1 types.Spans, overrideSet2 types.Spans) types.Spans

func NewDosagesFSM(params DosageFSMParams) DosagesFSM {
	machineSet := []fsm.Machine{
		getDosageQuantityMachine(params.TextNumberSet, params.SoloTextSet),
	}

	return func(sent *types.Sentence, overrideSet1 types.Spans, overrideSet2 types.Spans) types.Spans {
		outSet := make(types.Spans, 0)

		tokenStartMap := make(map[int]int)

		overrideTokenMap1 := make(map[int32]types.HasSpan)
		overrideTokenMap2 := make(map[int32]types.HasSpan)
		overrideBeginTokenMap1 := make(map[int32]int32)
		overrideBeginTokenMap2 := make(map[int32]int32)
		for _, t := range overrideSet1 {
			key := t.GetSpan().Begin
			overrideTokenMap1[key] = t
		}

		for _, t := range overrideSet2 {
			key := t.GetSpan().Begin
			overrideTokenMap2[key] = t
		}

		// init states
		machineStates := make([]string, len(machineSet))
		for n := 0; n < len(machineSet); n++ {
			machineStates[n] = startState
		}

		overrideOn1 := false
		overrideOn2 := false
		var overrideEndOffset1 int32 = -1
		var overrideEndOffset2 int32 = -1
		var tokenOffset1 int32 = 0
		var tokenOffset2 int32 = 0
		var anchorKey1 int32 = 0
		var anchorKey2 int32 = 0

		tokens := sent.Tokens
		for i, token := range tokens {

			var tokenSpan types.HasSpan = token
			key := tokenSpan.GetSpan().Begin

			if overrideOn1 && overrideOn2 {
				if overrideEndOffset1 >= overrideEndOffset2 {
					overrideOn1 = false
				} else {
					overrideOn2 = false
				}
			}

			if overrideOn1 {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset1 {
					overrideBeginTokenMap1[anchorKey1] = tokenOffset1
					overrideOn1 = false
					overrideEndOffset1 = -1
				} else {
					tokenOffset1++
					continue
				}
			} else if overrideOn2 {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset2 {
					overrideBeginTokenMap2[anchorKey2] = tokenOffset2
					overrideOn2 = false
					overrideEndOffset2 = -1
				} else {
					tokenOffset2++
					continue
				}
			} else {

				if overToken, isOk := overrideTokenMap1[key]; isOk {
					anchorKey1 = key
					tokenSpan = overToken
					overrideOn1 = true
					overrideEndOffset1 = tokenSpan.GetSpan().End
					tokenOffset1 = 0
				}
				if overToken, isOk := overrideTokenMap2[key]; isOk {
					anchorKey2 = key
					tokenSpan = overToken
					overrideOn2 = true
					overrideEndOffset2 = tokenSpan.GetSpan().End
					tokenOffset2 = 0
				}
			}

			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
					tokenOffset1 = 0
					tokenOffset2 = 0
				}

				if currentState == endState || currentState == ntEndState || currentState == ntFalseTermState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						var tokenMap1 int32 = 0
						var tokenMap2 int32 = 0

						lookUpOffset := tokens[tokenStartIndex]

						if offSet, isOk2 := overrideBeginTokenMap1[lookUpOffset.GetSpan().Begin]; isOk2 {
							tokenMap1 = offSet + tokenMap1
						}
						if offSet, isOk2 := overrideBeginTokenMap2[lookUpOffset.GetSpan().Begin]; isOk2 {
							tokenMap2 = offSet + tokenMap2
						}

						globalOffset := tokenMap1 + tokenMap2
						tokenStartIndex += int(globalOffset)

						tokenStartIndex++
					}

					startToken := tokens[tokenStartIndex]
					if currentState == ntFalseTermState {
						startToken = tokens[tokenStartIndex+1]
					}

					endToken := tokenSpan

					if currentState == ntEndState {
						endToken = tokens[i-1]
						tok, isOk2 := endToken.(*types.Token)
						if isOk2 && tok.IsPunct {
							endToken = tokens[i-2]
						}
					}

					outToken := DosageToken{
						Span: types.Span{
							Begin: startToken.GetSpan().Begin,
							End:   endToken.GetSpan().End,
						},
					}

					tokText, isOk := outToken.GetTextFromSentence(sent)
					if !isOk {
						continue
					}

					outToken.Text = utils.GlobalStringStore().GetPointer(tokText)
					outSet = append(outSet, &outToken)
					machineStates[machineIdx] = startState
				}
			}
		}
		return outSet
	}
}

func getDosageQuantityMachine(textNumberSet map[string]bool, soloTextSet map[string]bool) fsm.Machine {
	strengthFormCondition := fsm.NewDisjointCondition(RouteCondition, FormCondition)
	numberTextCondition := fsm.NewWordSetCondition(textNumberSet)
	soloTextCondition := fsm.NewWordSetCondition(soloTextSet)
	decimalStart := fsm.NewDisjointCondition(NewIntegerValueCondition(0), fsm.NumberCondition)
	hyphenCondition := fsm.NewPunctuationValueCondition('-')
	leftParenCondition := fsm.NewPunctuationValueCondition('(')
	ofCondition := fsm.NewTextValueCondition("of")
	aCondition := fsm.NewTextValueCondition("a")
	routeFormCondition := fsm.NewDisjointCondition(RouteCondition, FormCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: soloTextCondition, Dst: endState},
			{Cond: fsm.NumberCondition, Dst: dosageState},
			{Cond: RangeStrengthCondition, Dst: dosageState},
			{Cond: FractionStrengthCondition, Dst: dosageState},
			{Cond: numberTextCondition, Dst: dosageState},
			{Cond: decimalStart, Dst: dosageState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		formState: []fsm.MachineRule{
			{Cond: numberTextCondition, Dst: ntFalseTermState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dosageState: []fsm.MachineRule{
			{Cond: strengthFormCondition, Dst: ntEndState},
			{Cond: hyphenCondition, Dst: hyphState},
			{Cond: leftParenCondition, Dst: leftParenState},
			{Cond: ofCondition, Dst: ofState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ofState: []fsm.MachineRule{
			{Cond: aCondition, Dst: aState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		aState: []fsm.MachineRule{
			{Cond: routeFormCondition, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphState: []fsm.MachineRule{
			{Cond: soloTextCondition, Dst: endState},
			{Cond: fsm.NumberCondition, Dst: numState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numState: []fsm.MachineRule{
			{Cond: routeFormCondition, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftParenState: []fsm.MachineRule{
			{Cond: routeFormCondition, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntFalseTermState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
