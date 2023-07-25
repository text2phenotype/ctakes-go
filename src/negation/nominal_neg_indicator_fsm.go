package negation

import (
	"text2phenotype.com/fdl/fsm"
)

func getNominalNegIndicatorMachine() fsm.Machine {

	const (
		// states
		startState = "START"
		endState   = "END"
		ntEndState = "NON TERMINAL END"

		negPrepState = "NEG_PREP"
		negDetState  = "NEG_DET"
		regNounState = "REG_NOUN"
	)

	negPrepC := fsm.NewWordSetCondition(getNegPrepositions())
	negDetC := fsm.NewWordSetCondition(getNegDeterminers())
	regNounC := fsm.NewWordSetCondition(getRegNouns())

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: negDetC, Dst: negDetState},
			{Cond: negPrepC, Dst: negPrepState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		negPrepState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: ntEndState},
		},
		negDetState: []fsm.MachineRule{
			{Cond: regNounC, Dst: regNounState},
			{Cond: fsm.AnyCondition, Dst: ntEndState},
		},
		regNounState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: ntEndState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: endState},
		},
	}
}
