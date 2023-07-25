package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type DrugChangeStatusParams struct {
	SingleStopWordSet        map[string]bool `drug:"single_stop_word_set.txt"`
	SingleStartWordSet       map[string]bool `drug:"single_start_word_set.txt"`
	SingleIncreaseWordSet    map[string]bool `drug:"single_increase_word_set.txt"`
	SingleDecreaseWordSet    map[string]bool `drug:"single_decrease_word_set.txt"`
	SingleNoChangeWordSet    map[string]bool `drug:"single_no_change_word_set.txt"`
	MultiThenWordSet         map[string]bool `drug:"multi_then_word_set.txt"`
	SingleChangeWordSet      map[string]bool `drug:"single_change_word_set.txt"`
	FirstStartDualWordSet    map[string]bool `drug:"first_start_dual_word_set.txt"`
	FirstStopDualWordSet     map[string]bool `drug:"first_stop_dual_word_set.txt"`
	FirstNoChangeDualWordSet map[string]bool `drug:"first_no_change_dual_word_set.txt"`
	FirstIncreaseDualWordSet map[string]bool `drug:"first_increase_dual_word_set.txt"`
	FirstDecreaseDualWordSet map[string]bool `drug:"first_decrease_dual_word_set.txt"`
	SecondDualWordSet        map[string]bool `drug:"second_dual_word_set.txt"`
	SecondDualFromWordSet    map[string]bool `drug:"second_dual_from_word_set.txt"`
	SecondOffDualWordSet     map[string]bool `drug:"second_off_dual_word_set.txt"`
	NoChangeWordSet          map[string]bool `drug:"no_change_word_set.txt"`
	ChangeWordSet            map[string]bool `drug:"change_word_set.txt"`
	SingleMaxWordSet         map[string]bool `drug:"single_max_word_set.txt"`
	FirstMaxDualWordSet      map[string]bool `drug:"first_max_dual_word_set.txt"`
	SecondMaxDualWordSet     map[string]bool `drug:"second_max_dual_word_set.txt"`
	SingleSumWordSet         map[string]bool
}

type DrugChangeStatus string

const (
	Start        DrugChangeStatus = "start"
	Stop         DrugChangeStatus = "stop"
	IncreaseFrom DrugChangeStatus = "increasefrom"
	DecreaseFrom DrugChangeStatus = "decreasefrom"
	Increase     DrugChangeStatus = "increase"
	Decrease     DrugChangeStatus = "decrease"
	NoChange     DrugChangeStatus = "noChange"
	Other        DrugChangeStatus = "change"
	Sum          DrugChangeStatus = "add"
	Max          DrugChangeStatus = "max imum"
)

type DrugChangeStatusToken struct {
	types.Span
	Status DrugChangeStatus
}

func (token DrugChangeStatusToken) GetSpan() *types.Span {
	return &token.Span
}

type DrugChangeStatusFSM func(sent *types.Sentence) types.Spans

func NewDrugChangeStatusFSM(params DrugChangeStatusParams) DrugChangeStatusFSM {

	startStatusMachineIndex := 0
	stopStatusMachineIndex := 1
	increaseStatusMachineIndex := 2
	decreaseStatusMachineIndex := 3
	noChangeStatusMachineIndex := 4
	changeStatusMachineIndex := 5
	sumStatusMachineIndex := 6
	maxStatusMachineIndex := 7
	increaseFromStatusMachineIndex := 8
	decreaseFromStatusMachineIndex := 9

	machineSet := map[int]fsm.Machine{
		startStatusMachineIndex: getStartStatusMachine(
			params.SingleStartWordSet,
			params.FirstStartDualWordSet,
			params.SecondDualWordSet,
			params.MultiThenWordSet),
		stopStatusMachineIndex: getStopStatusMachine(
			params.SingleStopWordSet,
			params.FirstStopDualWordSet,
			params.SecondOffDualWordSet,
			params.SecondDualWordSet,
			params.MultiThenWordSet),
		increaseStatusMachineIndex: getIncreaseStatusMachine(
			params.SingleIncreaseWordSet,
			params.FirstIncreaseDualWordSet,
			params.SecondDualWordSet,
			params.MultiThenWordSet),
		decreaseStatusMachineIndex: getDecreaseStatusMachine(
			params.SingleDecreaseWordSet,
			params.FirstDecreaseDualWordSet,
			params.SecondDualWordSet,
			params.MultiThenWordSet),
		noChangeStatusMachineIndex: getNoChangeStatusMachine(
			params.SingleNoChangeWordSet,
			params.FirstNoChangeDualWordSet,
			params.SecondDualWordSet,
			params.MultiThenWordSet,
			params.NoChangeWordSet,
			params.ChangeWordSet,
		),
		changeStatusMachineIndex: getChangeStatusMachine(
			params.SingleChangeWordSet,
			params.ChangeWordSet),
		sumStatusMachineIndex: getSumStatusMachine(
			params.SingleSumWordSet),
		maxStatusMachineIndex: getMaximumStatusMachine(
			params.SingleMaxWordSet,
			params.FirstMaxDualWordSet,
			params.SecondMaxDualWordSet),
		increaseFromStatusMachineIndex: getIncreaseFromAndTheStatusMachine(
			params.FirstIncreaseDualWordSet,
			params.SecondDualFromWordSet,
			params.MultiThenWordSet),
		decreaseFromStatusMachineIndex: getDecreaseFromAndTheStatusMachine(
			params.FirstDecreaseDualWordSet,
			params.SecondDualFromWordSet,
			params.MultiThenWordSet),
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
				}

				if currentState == endState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					status := NoChange
					switch machineIdx {
					case startStatusMachineIndex:
						status = Start
					case stopStatusMachineIndex:
						status = Stop
					case increaseStatusMachineIndex:
						status = Increase
					case decreaseStatusMachineIndex:
						status = Decrease
					case noChangeStatusMachineIndex:
						status = NoChange
					case changeStatusMachineIndex:
						status = Other
					case sumStatusMachineIndex:
						status = Sum
					case maxStatusMachineIndex:
						status = Max
					case increaseFromStatusMachineIndex:
						status = IncreaseFrom
					case decreaseFromStatusMachineIndex:
						status = DecreaseFrom
					}

					outToken := DrugChangeStatusToken{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   token.GetSpan().End,
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

func getMaximumStatusMachine(
	singleMaxWordSet map[string]bool,
	firstMaxDualWordSet map[string]bool,
	secondMaxDualWordSet map[string]bool,
) fsm.Machine {

	singleMaxWordSetCondition := fsm.NewContainsSetTextValueCondition(singleMaxWordSet)
	firstMaxDualWordSetCondition := fsm.NewContainsSetTextValueCondition(firstMaxDualWordSet)
	secondMaxDualWordSetCondition := fsm.NewContainsSetTextValueCondition(secondMaxDualWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: singleMaxWordSetCondition, Dst: endState},
			{Cond: firstMaxDualWordSetCondition, Dst: foundDualFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		foundDualFirstState: []fsm.MachineRule{
			{Cond: secondMaxDualWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getChangeStatusMachine(
	singleChangeWordSet map[string]bool,
	changeWordSet map[string]bool,
) fsm.Machine {

	singleChangeWordSetCondition := fsm.NewContainsSetTextValueCondition(singleChangeWordSet)
	changeWordSetCondition := fsm.NewContainsSetTextValueCondition(changeWordSet)
	followedCondition := fsm.NewTextValueCondition("followed")
	byCondition := fsm.NewTextValueCondition("by")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: singleChangeWordSetCondition, Dst: endState},
			{Cond: followedCondition, Dst: byState},
			{Cond: changeWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		byState: []fsm.MachineRule{
			{Cond: byCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getSumStatusMachine(
	singleSumWordSet map[string]bool,
) fsm.Machine {

	singleSumWordSetCondition := fsm.NewContainsSetTextValueCondition(singleSumWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: singleSumWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getStartStatusMachine(
	singleStartWordSet map[string]bool,
	firstStartDualWordSet map[string]bool,
	secondDualWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {

	soloCondition := fsm.NewContainsSetTextValueCondition(singleStartWordSet)
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstStartDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)
	sectionCondition := fsm.NewTextValueCondition("section")
	sectionBracket := fsm.NewPunctuationValueCondition('[')
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: sectionBracket, Dst: beginEndState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: StrengthCondition, Dst: leftStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		beginEndState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endEndState: []fsm.MachineRule{
			{Cond: sectionCondition, Dst: leftStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getNoChangeStatusMachine(
	singleNoChangeWordSet map[string]bool,
	firstNoChangeDualWordSet map[string]bool,
	secondDualWordSet map[string]bool,
	multiThenWordSet map[string]bool,
	noChangeWordSet map[string]bool,
	changeWordSet map[string]bool,
) fsm.Machine {

	soloCondition := fsm.NewContainsSetTextValueCondition(singleNoChangeWordSet)
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstNoChangeDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)
	noChangeWordSetCondition := fsm.NewContainsSetTextValueCondition(noChangeWordSet)
	changeWordSetCondition := fsm.NewContainsSetTextValueCondition(changeWordSet)
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: StrengthCondition, Dst: leftStatusState},
			{Cond: noChangeWordSetCondition, Dst: dualWordState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		dualWordState: []fsm.MachineRule{
			{Cond: changeWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getIncreaseStatusMachine(
	singleIncreaseWordSet map[string]bool,
	firstIncreaseDualWordSet map[string]bool,
	secondDualWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {

	soloCondition := fsm.NewContainsSetTextValueCondition(singleIncreaseWordSet)
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstIncreaseDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getIncreaseFromAndTheStatusMachine(
	firstIncreaseDualWordSet map[string]bool,
	secondDualFromWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {

	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstIncreaseDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualFromWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getDecreaseStatusMachine(
	singleDecreaseWordSet map[string]bool,
	firstDecreaseDualWordSet map[string]bool,
	secondDualWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {
	soloCondition := fsm.NewContainsSetTextValueCondition(singleDecreaseWordSet)
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstDecreaseDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getDecreaseFromAndTheStatusMachine(
	firstDecreaseDualWordSet map[string]bool,
	secondDualFromWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstDecreaseDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualFromWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getStopStatusMachine(
	singleStopWordSet map[string]bool,
	firstStopDualWordSet map[string]bool,
	secondOffDualWordSet map[string]bool,
	secondDualWordSet map[string]bool,
	multiThenWordSet map[string]bool,
) fsm.Machine {
	soloCondition := fsm.NewContainsSetTextValueCondition(singleStopWordSet)
	firstDualCondition := fsm.NewContainsSetTextValueCondition(firstStopDualWordSet)
	secondOffDualCondition := fsm.NewContainsSetTextValueCondition(secondOffDualWordSet)
	secondDualCondition := fsm.NewContainsSetTextValueCondition(secondDualWordSet)
	thenCondition := fsm.NewContainsSetTextValueCondition(multiThenWordSet)

	sectionBracket := fsm.NewPunctuationValueCondition('[')
	sectionCondition := fsm.NewTextValueCondition("section")
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: thenCondition, Dst: thenStatusState},
			{Cond: soloCondition, Dst: endState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: sectionBracket, Dst: beginEndState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: StrengthCondition, Dst: leftStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		thenStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: sectionStatusState},
			{Cond: firstDualCondition, Dst: sectionStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sectionStatusState: []fsm.MachineRule{
			{Cond: secondDualCondition, Dst: endState},
			{Cond: secondOffDualCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		beginEndState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endEndState: []fsm.MachineRule{
			{Cond: sectionCondition, Dst: leftStatusState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftStatusState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
