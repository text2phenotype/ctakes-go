package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type FractionStrengthFSMParams struct {
	TextNumeratorSet   map[string]bool `drug:"text_numerator_set.txt"`
	TextDenominatorSet map[string]bool `drug:"text_denominator_set.txt"`
}

type FractionStrengthToken struct {
	types.Span
}

func (token FractionStrengthToken) GetSpan() *types.Span {
	return &token.Span
}

type FractionStrengthFSM func(sent *types.Sentence, overrideSet types.Spans) types.Spans

func NewFractionStrengthFSM(params FractionStrengthFSMParams) FractionStrengthFSM {

	machineSet := []fsm.Machine{
		getStrengthSlashMachine(params.TextNumeratorSet),
		getStandardMachine(
			params.TextNumeratorSet,
			params.TextDenominatorSet),
	}

	return func(sent *types.Sentence, overrideSet types.Spans) types.Spans {

		fractionSet := make(types.Spans, 0)

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

					outToken := FractionStrengthToken{
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
					fractionSet = append(fractionSet, &outToken)
					machineStates[machineIdx] = startState
				}
			}
		}
		return fractionSet
	}
}

func getStrengthSlashMachine(textNumeratorSet map[string]bool) fsm.Machine {
	leftContainsShortDose := fsm.NewContainsSetTextValueCondition(textNumeratorSet)
	fslashCondition := fsm.NewPunctuationValueCondition('/')
	containsdotCondition := fsm.NewPunctuationValueCondition('.')
	hyphenCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: numeratorLeftState},
			{Cond: leftContainsShortDose, Dst: numeratorLeftState},
			//{Cond: DecimalCondition, Dst: numeratorLeftState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numeratorLeftState: []fsm.MachineRule{
			{Cond: containsdotCondition, Dst: dotLeftState},
			{Cond: fslashCondition, Dst: fslashState},
			{Cond: hyphenCondition, Dst: numeratorRightState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dotLeftState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: hyphenState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphenState: []fsm.MachineRule{
			{Cond: hyphenCondition, Dst: numeratorRightState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numeratorRightState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dotRightState: []fsm.MachineRule{
			{Cond: containsdotCondition, Dst: fslashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		decPartNumState: []fsm.MachineRule{
			{Cond: fslashCondition, Dst: fslashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		fslashState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getStandardMachine(textNumeratorSet map[string]bool, textDenominatorSet map[string]bool) fsm.Machine {
	textNumeratorCondition := fsm.NewContainsSetTextValueCondition(textNumeratorSet)
	textDenominatorCondition := fsm.NewContainsSetTextValueCondition(textDenominatorSet)
	fslashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: numeratorNumState},
			{Cond: textNumeratorCondition, Dst: numeratorTextState},
			{Cond: textDenominatorCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numeratorNumState: []fsm.MachineRule{
			{Cond: fslashCondition, Dst: fslashState},
			{Cond: textDenominatorCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		fslashState: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numeratorTextState: []fsm.MachineRule{
			{Cond: textDenominatorCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
