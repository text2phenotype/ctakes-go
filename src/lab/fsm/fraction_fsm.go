package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type FractionFSM func(sent types.Sentence) []types.Token

func NewFractionFSM() FractionFSM {
	textNumeratorSet := map[string]bool{
		"one":   true,
		"two":   true,
		"three": true,
		"four":  true,
		"five":  true,
		"six":   true,
		"seven": true,
		"eight": true,
		"nine":  true,
		"ten":   true,
	}

	textDenominatorSet := map[string]bool{
		"half":     true,
		"halfs":    true,
		"third":    true,
		"thirds":   true,
		"fourth":   true,
		"fourths":  true,
		"fifth":    true,
		"fifths":   true,
		"sixth":    true,
		"sixths":   true,
		"seventh":  true,
		"sevenths": true,
		"eighth":   true,
		"eighths":  true,
		"nineths":  true,
		"nineth":   true,
		"tenth":    true,
		"tenths":   true,
	}

	machines := []fsm.Machine{getFractionMachine(textNumeratorSet, textDenominatorSet)}

	return func(sent types.Sentence) []types.Token {
		outSet := make([]types.Token, 0, len(sent.Tokens))

		const (
			Start = "START"
			End   = "END"
		)

		tokenStartMap := make(map[int]int)

		// init states
		machineStates := make([]string, len(machines))
		for n := 0; n < len(machines); n++ {
			machineStates[n] = Start
		}

		tokens := sent.Tokens
		for i, token := range tokens {
			for machineIdx, machine := range machines {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(token, currentState)
				machineStates[machineIdx] = currentState

				if currentState == Start {
					tokenStartMap[machineIdx] = i
				}

				if currentState == End {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					outToken := types.Token{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   token.GetSpan().End,
						},
						IsNumber: true,
					}

					tokText, isOk := outToken.GetTextFromSentence(&sent)
					if !isOk {
						continue
					}

					outToken.Shape = types.GetShape(tokText)
					outToken.Text = utils.GlobalStringStore().GetPointer(tokText)
					outSet = append(outSet, outToken)
					machineStates[machineIdx] = Start
				}
			}
		}

		return outSet
	}
}

/*
Gets a finite state machine that detects the following:
	- 1/2
	- half
	- one half
	- 1 half
*/
func getFractionMachine(textNumeratorSet map[string]bool, textDenominatorSet map[string]bool) fsm.Machine {
	// states
	const (
		Start         = "START"
		NumeratorNum  = "NUMERATOR_NUM"
		ForwardSlash  = "FORWARD_SLASH"
		NumeratorText = "NUMERATOR_TEXT"
		End           = "END"
	)

	fslashCondition := fsm.NewPunctuationValueCondition('/')
	textNumeratorCondition := fsm.NewWordSetCondition(textNumeratorSet)
	textDenominatorCondition := fsm.NewWordSetCondition(textDenominatorSet)

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: NumeratorNum},
			{Cond: textNumeratorCondition, Dst: NumeratorText},
			{Cond: textDenominatorCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		NumeratorNum: []fsm.MachineRule{
			{Cond: fslashCondition, Dst: ForwardSlash},
			{Cond: textDenominatorCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		ForwardSlash: []fsm.MachineRule{
			{Cond: fsm.NumberCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		NumeratorText: []fsm.MachineRule{
			{Cond: textDenominatorCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}
