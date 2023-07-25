package negation

import (
	"text2phenotype.com/fdl/fsm"
)

func getAspectualNegIndicatorMachine() fsm.Machine {

	const (
		// states
		startState      = "START"
		endState        = "END"
		anyState        = "ANY"
		ntEndState      = "NON TERMINAL END"
		regModalState   = "REG_MODAL"
		negPartState    = "NEG_PART"
		negVerbState    = "NEG_VERB"
		negCollocState  = "NEG_COLLOC"
		negColPartState = "NEG_COLPART"
	)

	regModalC := fsm.NewWordSetCondition(getModalVerbs())
	negPartC := fsm.NewWordSetCondition(getNegParticles())
	regVerbC := fsm.NewWordSetCondition(getRegVerbs())
	negVerbC := fsm.NewWordSetCondition(getNegVerbs())
	negDetC := fsm.NewWordSetCondition(getNegDeterminers())
	negCollocC := fsm.NewWordSetCondition(getNegColloc())
	negColPartC := fsm.NewWordSetCondition(getNegColPart())
	notCollocC := fsm.NewNegateCondition(negCollocC)
	c := fsm.NewDisjointCondition(negPartC, negDetC)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: negVerbC, Dst: negVerbState},
			{Cond: negCollocC, Dst: negCollocState},
			{Cond: fsm.NewDisjointCondition(regModalC, regVerbC), Dst: regModalState},
			{Cond: c, Dst: negPartState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		regModalState: []fsm.MachineRule{
			{Cond: negCollocC, Dst: negCollocState},
			{Cond: c, Dst: negPartState},
			{Cond: fsm.AnyCondition, Dst: anyState},
		},
		negCollocState: []fsm.MachineRule{
			{Cond: negColPartC, Dst: negColPartState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		negColPartState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: ntEndState},
		},
		anyState: []fsm.MachineRule{
			{Cond: c, Dst: negPartState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		negPartState: []fsm.MachineRule{
			{Cond: notCollocC, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		negVerbState: []fsm.MachineRule{
			{Cond: notCollocC, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: endState},
		},
	}
}
