package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type DecimalStrengthToken struct {
	types.Span
}

func (token DecimalStrengthToken) GetSpan() *types.Span {
	return &token.Span
}

type DecimalStrengthFSM func(sent *types.Sentence) types.Spans

func NewDecimalStrengthFSM() DecimalStrengthFSM {

	machineSet := []fsm.Machine{getDecimalStrengthMachine()}

	return func(sent *types.Sentence) types.Spans {
		outSet := make(types.Spans, 0)

		tokenStartMap := make(map[int]int)

		// init states
		machineStates := make([]string, len(machineSet))
		for n := 0; n < len(machineSet); n++ {
			machineStates[n] = startState
		}

		tokens := sent.Tokens
		for i, token := range tokens {
			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(token, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
				}

				if currentState == endState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					outToken := DecimalStrengthToken{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   token.GetSpan().End,
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

func getDecimalStrengthMachine() fsm.Machine {

	intCondition := fsm.NewIntegerValueCondition(0)
	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: intCondition, Dst: zeroNumState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		zeroNumState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: fractionTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		fractionTextState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: dashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dashState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
