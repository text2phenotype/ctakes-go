package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type RangeStrengthFSMParams struct {
	TextNumberSet map[string]bool `drug:"text_number_set.txt"`
	RangeSet      map[string]bool `drug:"range_set.txt"`
	HyphenatedSet map[string]bool `drug:"hyphenated_set.txt"`
}

type RangeStrengthToken struct {
	types.Span
}

func (token RangeStrengthToken) GetSpan() *types.Span {
	return &token.Span
}

type RangeStrengthFSM func(sent *types.Sentence) types.Spans

func NewRangeStrengthFSM(params RangeStrengthFSMParams) RangeStrengthFSM {
	machineSet := []fsm.Machine{
		getDashMachine(params.TextNumberSet, params.HyphenatedSet),
		getDotDashMachine(),
		getDashDashMachine(params.TextNumberSet, params.RangeSet),
	}

	return func(sent *types.Sentence) types.Spans {
		fractionSet := make(types.Spans, 0)

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

					outToken := RangeStrengthToken{
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

					fractionSet = append(fractionSet, &outToken)
					machineStates[machineIdx] = startState
				}
			}
		}
		return fractionSet
	}
}

/*
Gets a finite state machine that detects the following:
	25.4-30.4
	32.1-three
*/
func getDashDashMachine(textNumberSet map[string]bool, rangeSet map[string]bool) fsm.Machine {
	textNumberSetCondition := fsm.NewContainsSetTextValueCondition(textNumberSet)
	rangeSetCondition := fsm.NewContainsSetTextValueCondition(rangeSet)
	hyphenCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: {
			{Cond: textNumberSetCondition, Dst: leftNumTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftNumTextState: {
			{Cond: hyphenCondition, Dst: dash2State},
			{Cond: rangeSetCondition, Dst: rightNumTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dash2State: {
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: textNumberSetCondition, Dst: endState},
			{Cond: rangeSetCondition, Dst: middleDash},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleDash: {
			{Cond: hyphenCondition, Dst: dashAnotherState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightNumTextState: {
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: textNumberSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dashAnotherState: {
			{Cond: textNumberSetCondition, Dst: endState},
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

/*
Gets a finite state machine that detects the following:
	25.4-30.4
	32.1-three
*/
func getDashMachine(textNumberSet map[string]bool, hyphenatedSet map[string]bool) fsm.Machine {
	textNumberSetCondition := fsm.NewContainsSetTextValueCondition(textNumberSet)
	hyphenatedSetCondition := fsm.NewContainsSetTextValueCondition(hyphenatedSet)
	hyphenCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: {
			{Cond: fsm.NumberCondition, Dst: leftNumIntegerState},
			{Cond: hyphenatedSetCondition, Dst: endState},
			{Cond: textNumberSetCondition, Dst: leftNumTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftNumIntegerState: {
			{Cond: hyphenCondition, Dst: dashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftNumTextState: {
			{Cond: hyphenCondition, Dst: dashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dashState: {
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: textNumberSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dash1State: {
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: textNumberSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

/*
Gets a finite state machine that detects the following:
	250-300
	I-IV
	two-three
	two-to-three
*/
func getDotDashMachine() fsm.Machine {
	hyphenCondition := fsm.NewPunctuationValueCondition('-')
	dotCondition := fsm.NewPunctuationValueCondition('.')

	return fsm.Machine{
		startState: {
			{Cond: fsm.NumberCondition, Dst: leftNumIntegerState},
			{Cond: RangeCondition, Dst: leftNumIntegerState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftNumIntegerState: {
			{Cond: dotCondition, Dst: dotState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dotState: {
			{Cond: fsm.NumberCondition, Dst: decPartNumState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		decPartNumState: {
			{Cond: hyphenCondition, Dst: dashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dashState: {
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: FractionStrengthCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: {
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
