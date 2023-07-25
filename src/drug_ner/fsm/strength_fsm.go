package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type StrengthFSMParams struct {
	NumberTextSet map[string]bool `drug:"number_text_set.txt"`
}

type StrengthToken struct {
	types.Span
}

func (token StrengthToken) GetSpan() *types.Span {
	return &token.Span
}

type StrengthFSM func(sent *types.Sentence, overrideSet1 types.Spans, overrideSet2 types.Spans) types.Spans

func NewStrengthFSM(params StrengthFSMParams) StrengthFSM {
	machineSet := []fsm.Machine{getStrengthMachine(params.NumberTextSet)}

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

				if currentState == endState || currentState == ntEndState || currentState == ntEndHyphState {
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
					endToken := tokenSpan

					if currentState != endState && i > 0 {
						if currentState != ntEndHyphState {
							endToken = tokens[i-1]
						} else {
							endToken = tokens[i-2]
						}
					}

					outToken := StrengthToken{
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

func getStrengthMachine(numberTextSet map[string]bool) fsm.Machine {

	numberTextSetCondition := fsm.NewWordSetCondition(numberTextSet)
	nonSlashCondition := fsm.NewNegateCondition(fsm.NewPunctuationValueCondition('/'))
	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: RangeStrengthCondition, Dst: endState},
			{Cond: FractionStrengthCondition, Dst: dateState},
			{Cond: fsm.NumberCondition, Dst: connectState},
			//{Cond: DecimalCondition, Dst: connectState},
			{Cond: numberTextSetCondition, Dst: connectState},
			{Cond: StrengthUnitCombinedCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dateState: []fsm.MachineRule{
			{Cond: nonSlashCondition, Dst: connectState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		connectState: []fsm.MachineRule{
			{Cond: StrengthUnitCondition, Dst: ntEndState},
			{Cond: StrengthUnitCombinedCondition, Dst: endState},
			{Cond: dashCondition, Dst: unitState},
			{Cond: dotCondition, Dst: decimalState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		decimalState: []fsm.MachineRule{
			{Cond: StrengthUnitCondition, Dst: ntEndState},
			{Cond: StrengthUnitCombinedCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		unitState: []fsm.MachineRule{
			{Cond: StrengthUnitCondition, Dst: ntEndHyphState},
			{Cond: StrengthUnitCombinedCondition, Dst: endState},
			{Cond: dashCondition, Dst: unitState},
			{Cond: fsm.NumberCondition, Dst: complexState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		complexState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphenState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphenState: []fsm.MachineRule{
			{Cond: StrengthUnitCondition, Dst: ntEndHyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndHyphState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
