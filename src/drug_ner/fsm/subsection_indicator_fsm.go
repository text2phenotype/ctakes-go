package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type SubSectionIndicatorFSMParams struct {
	ProbableSubBeginSet  map[string]bool `drug:"probable_sub_begin_set.txt"`
	ProbableSubNextSet   map[string]bool `drug:"probable_sub_next_set.txt"`
	ProbableSubEndSet    map[string]bool `drug:"probable_sub_end_set.txt"`
	HistorySubBeginSet   map[string]bool `drug:"history_sub_begin_set.txt"`
	HistorySubNextSet    map[string]bool `drug:"history_sub_next_set.txt"`
	HistorySubMidSet     map[string]bool `drug:"history_sub_mid_set.txt"`
	ConfirmedSubBeginSet map[string]bool `drug:"confirmed_sub_begin_set.txt"`
	ConfirmedSubNextSet  map[string]bool `drug:"confirmed_sub_next_set.txt"`
	MiddleWordSet        map[string]bool `drug:"middle_word_set.txt"`
}

type SubSectionIndicatorStatus byte

const (
	ConfirmedStatus     SubSectionIndicatorStatus = 0
	HistoryStatus       SubSectionIndicatorStatus = 1
	FamilyHistoryStatus SubSectionIndicatorStatus = 2
	ProbableStatus      SubSectionIndicatorStatus = 3
)

type SubSectionIndicator struct {
	types.Span
	Status SubSectionIndicatorStatus
}

func (token SubSectionIndicator) GetSpan() *types.Span {
	return &token.Span
}

type SubSectionIndicatorFSM func(sent *types.Sentence) types.Spans

func NewSubSectionIndicatorFSM(params SubSectionIndicatorFSMParams) SubSectionIndicatorFSM {
	subSectionIDProbableMachine := getProbableSubSectionMachine(
		params.ProbableSubBeginSet,
		params.ProbableSubNextSet,
		params.MiddleWordSet,
		params.ProbableSubEndSet)
	subSectionIDHistoryMachine := getHistorySubSectionMachine(
		params.HistorySubBeginSet,
		params.HistorySubMidSet,
		params.HistorySubNextSet)
	subSectionIDConfirmMachine := getConfirmSubSectionMachine(
		params.ConfirmedSubBeginSet,
		params.ConfirmedSubNextSet)

	const ProbableMachineIndex = 0
	const HistoryMachineIndex = 1
	const ConfirmMachineIndex = 2

	machineSet := map[int]fsm.Machine{
		ProbableMachineIndex: subSectionIDProbableMachine,
		HistoryMachineIndex:  subSectionIDHistoryMachine,
		ConfirmMachineIndex:  subSectionIDConfirmMachine,
	}

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
				} else if currentState == endState || currentState == ntEndState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					endToken := token
					if currentState == ntEndState {
						endToken = tokens[i-1]
					}

					status := FamilyHistoryStatus
					if machineIdx == ProbableMachineIndex {
						status = ProbableStatus
					} else if machineIdx == HistoryMachineIndex {
						status = HistoryStatus
					} else if machineIdx == ConfirmMachineIndex {
						status = ConfirmedStatus
					}

					outToken := SubSectionIndicator{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   endToken.GetSpan().End,
						},
						Status: status,
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

func getHistorySubSectionMachine(historySubBeginSet map[string]bool, historySubMidSet map[string]bool, historySubNextSet map[string]bool) fsm.Machine {
	subFirstBegin := fsm.NewContainsSetTextValueCondition(historySubBeginSet)
	subFirstMid := fsm.NewContainsSetTextValueCondition(historySubMidSet)
	subFirstNext := fsm.NewContainsSetTextValueCondition(historySubNextSet)
	colonCondition := fsm.NewPunctuationValueCondition(':')

	return fsm.Machine{
		startState: {
			{Cond: subFirstBegin, Dst: medState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		medState: {
			{Cond: subFirstNext, Dst: endState},
			{Cond: subFirstMid, Dst: midWordState},
			{Cond: colonCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		midWordState: {
			{Cond: subFirstNext, Dst: endState},
			{Cond: colonCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getProbableSubSectionMachine(
	probableSubBeginSet map[string]bool,
	probableSubNextSet map[string]bool,
	middleWordSet map[string]bool,
	probableSubEndSet map[string]bool,
) fsm.Machine {

	subFirstBegin := fsm.NewContainsSetTextValueCondition(probableSubBeginSet)
	subFirstNext := fsm.NewContainsSetTextValueCondition(probableSubNextSet)
	subFirstEnd := fsm.NewContainsSetTextValueCondition(probableSubEndSet)
	middleCondition := fsm.NewContainsSetTextValueCondition(middleWordSet)

	return fsm.Machine{
		startState: {
			{Cond: subFirstBegin, Dst: medState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		medState: {
			{Cond: subFirstNext, Dst: endState},
			{Cond: middleCondition, Dst: endWordState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endWordState: {
			{Cond: subFirstEnd, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getConfirmSubSectionMachine(confirmedSubBeginSet map[string]bool, confirmedSubNextSet map[string]bool) fsm.Machine {
	subFirstBegin := fsm.NewContainsSetTextValueCondition(confirmedSubBeginSet)
	subFirstNext := fsm.NewContainsSetTextValueCondition(confirmedSubNextSet)
	dotCondition := fsm.NewPunctuationValueCondition('.')
	pCondition := fsm.NewTextValueCondition("p")
	rCondition := fsm.NewTextValueCondition("r")
	nCondition := fsm.NewTextValueCondition("n")

	return fsm.Machine{
		startState: {
			{Cond: subFirstBegin, Dst: medState},
			{Cond: pCondition, Dst: firstDotState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotState: {
			{Cond: dotCondition, Dst: rState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rState: {
			{Cond: rCondition, Dst: secondDotState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotState: {
			{Cond: dotCondition, Dst: nState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		nState: {
			{Cond: nCondition, Dst: thirdDotState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thirdDotState: {
			{Cond: dotCondition, Dst: medState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		medState: {
			{Cond: subFirstNext, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
