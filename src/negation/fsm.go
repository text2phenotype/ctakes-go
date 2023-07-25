package negation

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
)

type Condition func(token *string) bool
type PolarityFSM func(tokens []*types.Token) bool

func NewPolarityFSM() PolarityFSM {

	const (
		// states
		startState = "START"
		endState   = "END"
		ntEndState = "NON TERMINAL END"
	)

	machines := []fsm.Machine{
		getAspectualNegIndicatorMachine(),
		getNominalNegIndicatorMachine(),
		getAdjNegIndicatorMachine(),
		getCorrectionAdjNegIndicatorMachine(),
	}

	return func(tokens []*types.Token) bool {
		tokenStartMap := make(map[int]int)

		// init states
		machineStates := make([]string, len(machines))
		for n := 0; n < len(machines); n++ {
			machineStates[n] = startState
		}

		for i, token := range tokens {
			for machineIdx, machine := range machines {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(token, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
				}

				if currentState == endState || currentState == ntEndState {
					return true
				}
			}
		}
		return false
	}
}
