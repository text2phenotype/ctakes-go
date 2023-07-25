package negation

import (
	"text2phenotype.com/fdl/fsm"
)

func getAdjNegIndicatorMachine() fsm.Machine {

	const (
		// states
		startState = "START"
		endState   = "END"
		ntEndState = "NON TERMINAL END"

		regPrepState = "REG_PREP"
		negAdjState  = "NEG_ADJ"
	)

	regPrepC := fsm.NewWordSetCondition(getNegPrepositions())
	negAdjC := fsm.NewWordSetCondition(getNegAdjectives())

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: negAdjC, Dst: negAdjState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		negAdjState: []fsm.MachineRule{
			{Cond: regPrepC, Dst: regPrepState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		regPrepState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: ntEndState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: endState},
		},
	}
}

func getCorrectionAdjNegIndicatorMachine() fsm.Machine {

	const (
		// states
		startState = "START"
		endState   = "END"
	)

	negAdjC := fsm.NewWordSetCondition(getNegAdjectives())

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: negAdjC, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
