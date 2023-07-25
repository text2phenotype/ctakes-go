package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type StrengthUnitFSMParams struct {
	FullTextSet map[string]bool `drug:"full_text_set.txt"`
}

type StrengthUnitToken struct {
	types.Span
}

func (token StrengthUnitToken) GetSpan() *types.Span {
	return &token.Span
}

type StrengthUnitCombinedToken struct {
	types.Span
}

func (token StrengthUnitCombinedToken) GetSpan() *types.Span {
	return &token.Span
}

type StrengthUnitFSM func(sent *types.Sentence, overrideSet types.Spans) types.Spans

func NewStrengthUnitFSM(params StrengthUnitFSMParams) StrengthUnitFSM {
	strengthCombinedMachineIndex := 0
	strengthdMachineIndex := 1
	machineSet := map[int]fsm.Machine{
		strengthCombinedMachineIndex: getStrengthUnitMachine1(params.FullTextSet),
		strengthdMachineIndex:        getStrengthUnitMachine2(params.FullTextSet),
	}

	return func(sent *types.Sentence, overrideSet types.Spans) types.Spans {
		outSet := make(types.Spans, 0)

		tokenStartMap := make(map[int]int)

		overrideTokenMap := make(map[int32]types.HasSpan)
		overrideBeginTokenMap := make(map[int32]int32)
		for _, t := range overrideSet {
			key := t.GetSpan().Begin
			overrideTokenMap[key] = t
		}

		// init states
		machineStates := make([]string, len(machineSet))
		for n := 0; n < len(machineSet); n++ {
			machineStates[n] = startState
		}

		overrideOn := false
		var overrideEndOffset int32 = -1
		var tokenOffset int32 = 0
		var anchorKey int32 = 0

		tokens := sent.Tokens
		for i, token := range tokens {
			var tokenSpan types.HasSpan = token
			key := tokenSpan.GetSpan().Begin

			if overrideOn {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset {
					if tokenOffset > 0 {
						overrideBeginTokenMap[anchorKey] = tokenOffset
					}

					overrideOn = false
					overrideEndOffset = -1
				} else {
					tokenOffset++
					continue
				}
			} else {

				if overToken, isOk := overrideTokenMap[key]; isOk {
					anchorKey = key
					tokenSpan = overToken
					overrideOn = true
					overrideEndOffset = tokenSpan.GetSpan().End
					tokenOffset = 0
				}
			}

			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
					tokenOffset = 0
				}

				if currentState == endState || currentState == ntFalseTermState {

					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						var tokenMap int32 = 0

						lookUpOffset := tokens[tokenStartIndex]

						if offSet, isOk2 := overrideBeginTokenMap[lookUpOffset.GetSpan().Begin]; isOk2 {
							tokenMap = offSet + tokenMap
						}

						tokenStartIndex += int(tokenMap)

						tokenStartIndex++
					}

					startToken := tokens[tokenStartIndex]
					endToken := tokenSpan

					if currentState == ntFalseTermState {
						startToken = tokens[tokenStartIndex+1]
					}

					if machineIdx == strengthCombinedMachineIndex {
						outToken := StrengthUnitCombinedToken{
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
					} else {
						outToken := StrengthUnitToken{
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
					}

					machineStates[machineIdx] = startState
				}
			}
		}
		return outSet
	}
}

func getStrengthUnitMachine2(fullTextSet map[string]bool) fsm.Machine {

	fullTextSetCondition := fsm.NewWordSetCondition(fullTextSet)
	percentCondition := fsm.NewPunctuationValueCondition('%')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: fullTextSetCondition, Dst: endState},
			{Cond: dashCondition, Dst: unitState},
			{Cond: percentCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		unitState: []fsm.MachineRule{
			{Cond: fullTextSetCondition, Dst: ntFalseTermState},
			{Cond: percentCondition, Dst: endState},
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

func getStrengthUnitMachine1(fullTextSet map[string]bool) fsm.Machine {

	fullTextSetCondition := fsm.NewContainsSetTextValueCondition(fullTextSet)
	percentCondition := fsm.NewPunctuationValueCondition('%')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: fullTextSetCondition, Dst: endState},
			{Cond: dashCondition, Dst: unitState},
			{Cond: percentCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		unitState: []fsm.MachineRule{
			{Cond: fullTextSetCondition, Dst: ntFalseTermState},
			{Cond: percentCondition, Dst: endState},
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
