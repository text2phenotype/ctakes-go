package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"sort"
)

type DateFSM func(sent types.Sentence, overrideSet []types.Token) []types.Token

func NewDateFSM() DateFSM {
	fullMonthNameSet := map[string]bool{
		"january":   true,
		"february":  true,
		"march":     true,
		"april":     true,
		"may":       true,
		"june":      true,
		"july":      true,
		"august":    true,
		"september": true,
		"october":   true,
		"november":  true,
		"december":  true,
	}

	shortMonthNameSet := map[string]bool{
		"jan":  true,
		"feb":  true,
		"mar":  true,
		"apr":  true,
		"may":  true,
		"jun":  true,
		"jul":  true,
		"aug":  true,
		"sep":  true,
		"sept": true,
		"oct":  true,
		"nov":  true,
		"dec":  true,
	}

	machines := []fsm.Machine{
		getLongTextualDateMachine(fullMonthNameSet, shortMonthNameSet),
		getShortTextualDateMachine(fullMonthNameSet, shortMonthNameSet),
		getLongNumericDateMachine(),
		getShortNumericDateMachine(),
	}
	return func(sent types.Sentence, overrideSet []types.Token) []types.Token {
		outSet := make([]types.Token, 0, len(sent.Tokens))

		const (
			Start = "START"
			End   = "END"
		)

		tokenStartMap := make(map[int]int)

		overrideTokenMap := make(map[int32]types.HasSpan)
		for _, t := range overrideSet {
			key := t.GetSpan().Begin
			overrideTokenMap[key] = &t
		}

		// init states
		machineStates := make([]string, len(machines))
		for n := 0; n < len(machines); n++ {
			machineStates[n] = Start
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

			for machineIdx, machine := range machines {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
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
							End:   tokenSpan.GetSpan().End,
						},
						IsWord: true,
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

		if len(outSet) > 1 {
			sort.SliceStable(outSet, func(i, j int) bool {
				if outSet[i].Begin < outSet[j].Begin {
					return true
				}

				if outSet[i].Begin == outSet[j].Begin {
					return outSet[i].End < outSet[j].End
				}

				return outSet[i].End < outSet[j].End
			})

			deduplicatedOutSet := make([]types.Token, 0, len(outSet))
			for i := 0; i < len(outSet)-1; i++ {
				if outSet[i].End < outSet[i+1].Begin {
					deduplicatedOutSet = append(deduplicatedOutSet, outSet[i])
				}
			}

			deduplicatedOutSet = append(deduplicatedOutSet, outSet[len(outSet)-1])
			outSet = deduplicatedOutSet

		}
		return outSet
	}
}

/*
Gets a finite state machine that detects the following:
	October 15, 2002
	October 15 2002
	Oct 15, 2002
	Oct 15 2002
	Oct. 15, 2002
	Oct. 15 2002
*/
func getLongTextualDateMachine(fullMonthNameSet map[string]bool, shortMonthNameSet map[string]bool) fsm.Machine {
	// states
	const (
		Start      = "START"
		LongMonth  = "LONG_MONTH"
		ShortMonth = "SHORT_MONTH"
		Day        = "DAY"
		Dot        = "DOT"
		Comma      = "COMMA"
		End        = "END"
	)

	fullMonthCondition := fsm.NewWordSetCondition(fullMonthNameSet)
	shortMonthCondition := fsm.NewWordSetCondition(shortMonthNameSet)
	dayCondition := fsm.NewIntegerRangeCondition(1, 31)
	yearCondition := fsm.NewIntegerRangeCondition(1900, 2100)
	dotCondition := fsm.NewPunctuationValueCondition('.')
	commaCondition := fsm.NewPunctuationValueCondition(',')

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: fullMonthCondition, Dst: LongMonth},
			{Cond: shortMonthCondition, Dst: ShortMonth},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		LongMonth: []fsm.MachineRule{
			{Cond: dayCondition, Dst: Day},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		ShortMonth: []fsm.MachineRule{
			{Cond: dotCondition, Dst: Dot},
			{Cond: dayCondition, Dst: Day},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Dot: []fsm.MachineRule{
			{Cond: dayCondition, Dst: Day},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Day: []fsm.MachineRule{
			{Cond: commaCondition, Dst: Comma},
			{Cond: yearCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Comma: []fsm.MachineRule{
			{Cond: yearCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}

/*
Gets a finite state machine that detects the following:
	October 2002
	Oct. 2002
	Oct 2002
	October 15
	Oct 15
	Oct. 15
*/
func getShortTextualDateMachine(fullMonthNameSet map[string]bool, shortMonthNameSet map[string]bool) fsm.Machine {
	// states
	const (
		Start = "START"
		Month = "MONTH"
		Dot   = "DOT"
		End   = "END"
	)

	fullMonthCondition := fsm.NewWordSetCondition(fullMonthNameSet)
	shortMonthCondition := fsm.NewWordSetCondition(shortMonthNameSet)
	dayCondition := fsm.NewIntegerRangeCondition(1, 31)
	yearCondition := fsm.NewIntegerRangeCondition(1900, 2100)
	dotCondition := fsm.NewPunctuationValueCondition('.')

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: fullMonthCondition, Dst: Month},
			{Cond: shortMonthCondition, Dst: Month},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Month: []fsm.MachineRule{
			{Cond: dotCondition, Dst: Dot},
			{Cond: dayCondition, Dst: End},
			{Cond: yearCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Dot: []fsm.MachineRule{
			{Cond: dayCondition, Dst: End},
			{Cond: yearCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}

/*
Gets a finite state machine that detects the following:
	- 10/15/1994
	- 10-15-1994
	- 10/15/94
	- 10-15-94
*/
func getLongNumericDateMachine() fsm.Machine {

	// states
	const (
		Start           = "START"
		Month           = "MONTH"
		Day             = "DAY"
		FirstSeparator  = "SEPARATOR1"
		SecondSeparator = "SEPARATOR2"
		End             = "END"
	)

	monthCondition := fsm.NewIntegerRangeCondition(1, 12)
	dayCondition := fsm.NewIntegerRangeCondition(1, 31)
	longYearCondition := fsm.NewIntegerRangeCondition(1900, 2100)
	shortYearCondition := fsm.NewIntegerRangeCondition(00, 99)
	slashCondition := fsm.NewPunctuationValueCondition('/')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: monthCondition, Dst: Month},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Month: []fsm.MachineRule{
			{Cond: slashCondition, Dst: FirstSeparator},
			{Cond: dashCondition, Dst: FirstSeparator},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		FirstSeparator: []fsm.MachineRule{
			{Cond: dayCondition, Dst: Day},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Day: []fsm.MachineRule{
			{Cond: slashCondition, Dst: SecondSeparator},
			{Cond: dashCondition, Dst: SecondSeparator},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		SecondSeparator: []fsm.MachineRule{
			{Cond: longYearCondition, Dst: End},
			{Cond: shortYearCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}

/*
Gets a finite state machine that detects the following:
	- 10/15
	- 10-15
*/
func getShortNumericDateMachine() fsm.Machine {

	// states
	const (
		Start     = "START"
		Month     = "MONTH"
		Separator = "SEPARATOR"
		End       = "END"
	)

	monthCondition := fsm.NewIntegerRangeCondition(1, 12)
	dayCondition := fsm.NewIntegerRangeCondition(1, 31)
	slashCondition := fsm.NewPunctuationValueCondition('/')
	dashCondition := fsm.NewPunctuationValueCondition('-')

	return fsm.Machine{
		Start: []fsm.MachineRule{
			{Cond: monthCondition, Dst: Month},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Month: []fsm.MachineRule{
			{Cond: slashCondition, Dst: Separator},
			{Cond: dashCondition, Dst: Separator},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		Separator: []fsm.MachineRule{
			{Cond: dayCondition, Dst: End},
			{Cond: fsm.AnyCondition, Dst: Start},
		},
		End: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: Start},
		},
	}
}
