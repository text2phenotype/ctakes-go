package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type TimeFSMParams struct {
	DayNightSet map[string]bool `drug:"day_night_set.txt"`
}

type TimeToken struct {
	types.Span
}

func (token TimeToken) GetSpan() *types.Span {
	return &token.Span
}

type TimeFSM func(sent *types.Sentence) types.Spans

func NewTimeFSM(params TimeFSMParams) TimeFSM {

	machineSet := []fsm.Machine{
		getTimeMachine(params.DayNightSet),
		getTime24Machine(),
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

					outToken := TimeToken{
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

					outSet = append(outSet, &outToken)
					machineStates[machineIdx] = startState
				}
			}
		}
		return outSet
	}
}

func getTimeMachine(dayNightSet map[string]bool) fsm.Machine {

	const (
		hourNumState            = "HOUR_NUM"
		hourMinTextState        = "HOUR_MIN_TEXT"
		ampmTextWithPeriodState = "AM_PM_PERIOD_TEXT"
	)

	dayNightCondition := fsm.NewWordSetCondition(dayNightSet)
	hourNumCondition := fsm.NewIntegerRangeCondition(1, 12)
	hourMinCondition := NewHourMinuteCondition(1, 12, 0, 59)
	dayNightWithPeriodCondition := DayNightWordCondition
	closingPeriodCondition := fsm.NewPunctuationValueCondition('.')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: hourNumCondition, Dst: hourNumState},
			{Cond: hourMinCondition, Dst: hourMinTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hourMinTextState: []fsm.MachineRule{
			{Cond: dayNightCondition, Dst: endState},
			{Cond: dayNightWithPeriodCondition, Dst: ampmTextWithPeriodState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hourNumState: []fsm.MachineRule{
			{Cond: dayNightCondition, Dst: endState},
			{Cond: dayNightWithPeriodCondition, Dst: ampmTextWithPeriodState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ampmTextWithPeriodState: []fsm.MachineRule{
			{Cond: closingPeriodCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getTime24Machine() fsm.Machine {

	const (
		hourNumState   = "HOUR_NUM"
		separatorState = "SEPARATOR"
	)

	hourNumCondition := fsm.NewIntegerRangeCondition(0, 23)
	minuteNumCondition := fsm.NewIntegerRangeCondition(0, 59)
	separatorCondition := fsm.NewPunctuationValueCondition(':')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: hourNumCondition, Dst: hourNumState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hourNumState: []fsm.MachineRule{
			{Cond: separatorCondition, Dst: separatorState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		separatorState: []fsm.MachineRule{
			{Cond: minuteNumCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
