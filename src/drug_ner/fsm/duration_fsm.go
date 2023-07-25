package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type DurationFSMParams struct {
	MiddleNumericTermSet map[string]bool `drug:"middle_numeric_term_set.txt"`
	CombinedSet          map[string]bool `drug:"combined_set.txt"`
	SpecifiedWordSet     map[string]bool `drug:"specified_word_set.txt"`
	AppendWordSet        map[string]bool `drug:"append_word_set.txt"`
	PeriodSet            map[string]bool `drug:"period_set.txt"`
}

type DurationToken struct {
	types.Span
}

func (token DurationToken) GetSpan() *types.Span {
	return &token.Span
}

type DurationFSM func(sent *types.Sentence, overrideSet types.Spans) types.Spans

func NewDurationFSM(params DurationFSMParams) DurationFSM {

	machineSet := []fsm.Machine{
		geDurationMachine(
			params.MiddleNumericTermSet,
			params.PeriodSet,
			params.SpecifiedWordSet,
			params.CombinedSet),
		geDuration2ndMachine(
			params.MiddleNumericTermSet,
			params.PeriodSet,
			params.SpecifiedWordSet,
			params.AppendWordSet),
	}

	return func(sent *types.Sentence, overrideSet types.Spans) types.Spans {
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

					outToken := DurationToken{
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

func geDuration2ndMachine(
	middleNumericTermSet map[string]bool,
	periodSet map[string]bool,
	specifiedWordSet map[string]bool,
	appendWordSet map[string]bool,
) fsm.Machine {

	numericTextCondition := fsm.NewContainsSetTextValueCondition(middleNumericTermSet)
	periodCondition := fsm.NewContainsSetTextValueCondition(periodSet)
	specificWordCondition := fsm.NewContainsSetTextValueCondition(specifiedWordSet)
	containsAppendTermCondition := fsm.NewContainsSetTextValueCondition(appendWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: specificWordCondition, Dst: startAbbreviateState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startAbbreviateState: []fsm.MachineRule{
			{Cond: containsAppendTermCondition, Dst: middleTermState},
			{Cond: fsm.NumberCondition, Dst: finalTermState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleTermState: []fsm.MachineRule{
			{Cond: RangeStrengthCondition, Dst: anotherAppendState},
			{Cond: containsAppendTermCondition, Dst: finalTermState},
			{Cond: numericTextCondition, Dst: finalAppendState},
			{Cond: fsm.NumberCondition, Dst: anotherAppendState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		finalTermState: []fsm.MachineRule{
			{Cond: RangeStrengthCondition, Dst: finalTextState},
			{Cond: numericTextCondition, Dst: finalTextState},
			{Cond: fsm.NumberCondition, Dst: finalTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		finalAppendState: []fsm.MachineRule{
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		anotherAppendState: []fsm.MachineRule{
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		finalTextState: []fsm.MachineRule{
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func geDurationMachine(
	middleNumericTermSet map[string]bool,
	periodSet map[string]bool,
	specifiedWordSet map[string]bool,
	combinedSet map[string]bool,
) fsm.Machine {

	middleTextCondition := fsm.NewContainsSetTextValueCondition(middleNumericTermSet)
	periodCondition := fsm.NewContainsSetTextValueCondition(periodSet)
	specificWordCondition := fsm.NewContainsSetTextValueCondition(specifiedWordSet)
	combinedCondition := fsm.NewContainsSetTextValueCondition(combinedSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: specificWordCondition, Dst: leftAbbreviateState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: RangeStrengthCondition, Dst: middleTextState},
			{Cond: middleTextCondition, Dst: middleTextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: fsm.NumberCondition, Dst: middleTextState},
			{Cond: combinedCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleTextState: []fsm.MachineRule{
			{Cond: RangeStrengthCondition, Dst: lastTextState},
			{Cond: middleTextCondition, Dst: lastTextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: dashCondition, Dst: secondDashState},
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: middleTextCondition, Dst: middleTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDashState: []fsm.MachineRule{
			{Cond: middleTextCondition, Dst: lastTextState},
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: middleTextCondition, Dst: endState},
			{Cond: periodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
