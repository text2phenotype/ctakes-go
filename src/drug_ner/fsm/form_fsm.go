package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type FormFSMParams struct {
	FullTextSet map[string]bool `drug:"full_text_set.txt"`
}

type FormToken struct {
	types.Span
}

func (token FormToken) GetSpan() *types.Span {
	return &token.Span
}

type FormFSM func(sent *types.Sentence, overrideSet []types.HasSpan) types.Spans

func NewFormFSM(params FormFSMParams) FormFSM {
	machineSet := []fsm.Machine{getFormMachine(params.FullTextSet)}

	return func(sent *types.Sentence, overrideSet []types.HasSpan) types.Spans {
		outSet := make(types.Spans, 0)

		tokenStartMap := make(map[int]int)

		overrideTokenMap := make(map[int32]types.HasSpan)
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

		tokens := sent.Tokens
		for i, token := range tokens {
			var tokenSpan types.HasSpan = token
			key := tokenSpan.GetSpan().Begin

			if overrideOn {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset {
					overrideOn = false
					overrideEndOffset = -1
				} else {
					continue
				}
			} else {
				if overToken, isOk := overrideTokenMap[key]; isOk {
					tokenSpan = overToken
					overrideOn = true
					overrideEndOffset = tokenSpan.GetSpan().End
				}
			}

			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
				}

				if currentState == endState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					outToken := FormToken{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   tokenSpan.GetSpan().End,
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

func getFormMachine(fullTextSet map[string]bool) fsm.Machine {

	fullTextCondition := fsm.NewWordSetCondition(fullTextSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: fullTextCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
