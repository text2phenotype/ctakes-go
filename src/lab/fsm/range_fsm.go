package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type RangeFSM func(sent types.Sentence) []types.Token

func NewRangeFSM() RangeFSM {
	textNumberSet := map[string]bool{
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

	machines := []fsm.Machine{getRangesMachine(textNumberSet)}
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
	- 250-300
	- two-three
*/
func getRangesMachine(textNumberSet map[string]bool) fsm.Machine {
	// states
	const (
		Start  = "START"
		Number = "NUMBER"
		Dash   = "DASH"
		End    = "END"
	)

	textNumberCondition := fsm.NewWordSetCondition(textNumberSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: textNumberCondition, Dst: Number},
			{Cond: fsm.NumberCondition, Dst: Number},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Number: []fsm.MachineRule{
			{Cond: dashCondition, Dst: Dash},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Dash: []fsm.MachineRule{
			{Cond: textNumberCondition, Dst: End},
			{Cond: fsm.NumberCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}
