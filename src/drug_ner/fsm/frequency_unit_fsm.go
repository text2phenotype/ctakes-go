package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type FrequencyUnitFSMParams struct {
	DailyWordSet           map[string]bool `drug:"daily_word_set.txt"`
	PerDayWordSet          map[string]bool `drug:"per_day_word_set.txt"`
	PostEightWordSet       map[string]bool `drug:"post_eight_word_set.txt"`
	PostFourWordSet        map[string]bool `drug:"post_four_word_set.txt"`
	PostSixWordSet         map[string]bool `drug:"post_six_word_set.txt"`
	FourTimesPerDayWordSet map[string]bool `drug:"four_times_per_day_word_set.txt"`
	PerWeekWordSet         map[string]bool `drug:"per_week_word_set.txt"`
	TwiceADayWordSet       map[string]bool `drug:"twice_a_day_word_set.txt"`
	ThreeTimesADayWordSet  map[string]bool `drug:"three_times_a_day_word_set.txt"`
	SixTimesPerDayWordSet  map[string]bool `drug:"six_times_per_day_word_set.txt"`
	EveryOtherHourWordSet  map[string]bool `drug:"every_other_hour_word_set.txt"`
	DailySuffixSet         map[string]bool `drug:"daily_suffix_set.txt"`
	WeeklySuffixSet        map[string]bool `drug:"weekly_suffix_set.txt"`
	YearlySuffixSet        map[string]bool `drug:"yearly_suffix_set.txt"`
	HourlySuffixSet        map[string]bool `drug:"hourly_suffix_set.txt"`
	MonthlySuffixSet       map[string]bool `drug:"monthly_suffix_set.txt"`
	PrnWordSet             map[string]bool `drug:"prn_word_set.txt"`
	EveryOtherDayWordSet   map[string]bool `drug:"every_other_day_word_set.txt"`
}

type FrequencyUnitQuantity string

const (
	QuantityPrn           FrequencyUnitQuantity = "0.0"
	QuantityOne           FrequencyUnitQuantity = "1.0"
	QuantityTwo           FrequencyUnitQuantity = "2.0"
	QuantityThree         FrequencyUnitQuantity = "3.0"
	QuantityFour          FrequencyUnitQuantity = "4.0"
	QuantityFive          FrequencyUnitQuantity = "5.0"
	QuantitySix           FrequencyUnitQuantity = "6.0"
	QuantitySeven         FrequencyUnitQuantity = "7.0"
	QuantityEight         FrequencyUnitQuantity = "8.0"
	QuantityNine          FrequencyUnitQuantity = "9.0"
	QuantityTen           FrequencyUnitQuantity = "10.0"
	QuantityEleven        FrequencyUnitQuantity = "11.0"
	Quantity24            FrequencyUnitQuantity = "24.0"
	QuantityWeekly        FrequencyUnitQuantity = "0.14"
	QuantityBiweekly      FrequencyUnitQuantity = "0.07"
	QuantityMonthly       FrequencyUnitQuantity = "0.03"
	QuantityEveryOtherDay FrequencyUnitQuantity = "0.5"
	QuantityYearly        FrequencyUnitQuantity = "0.003"
)

type FrequencyUnitToken struct {
	types.Span
	Quantity FrequencyUnitQuantity
}

func (token FrequencyUnitToken) GetSpan() *types.Span {
	return &token.Span
}

type FrequencyUnitFSM func(sent *types.Sentence, overrideSet []types.HasSpan) types.Spans

func NewFrequencyUnitFSM(params FrequencyUnitFSMParams) FrequencyUnitFSM {
	DailyMachine := 0
	SixTimesADayMachine := 1
	FiveTimesADayMachine := 2
	ThreeTimesADayMachine := 3
	FourTimesADayMachine := 4
	EveryOtherHourMachine := 5
	EveryOtherDayMachine := 6
	TwiceADayMachine := 7
	DailySuffixMachine := 8
	WeeklyMachine := 9
	HourlySuffixMachine := 10
	WeeklySuffixMachine := 11
	MonthlySuffixMachine := 12
	YearlySuffixMachine := 13
	PrnMachine := 14

	machineSet := map[int]fsm.Machine{
		DailyMachine: getDailyMachine(
			params.DailyWordSet,
			params.PerDayWordSet),
		SixTimesADayMachine: getSixTimesADayMachine(
			params.SixTimesPerDayWordSet,
			params.PostFourWordSet,
			params.HourlySuffixSet),
		FiveTimesADayMachine: getFiveTimesADayMachine(
			params.SixTimesPerDayWordSet,
			params.HourlySuffixSet),
		ThreeTimesADayMachine: getThreeTimesADayMachine(
			params.ThreeTimesADayWordSet,
			params.PostEightWordSet,
			params.HourlySuffixSet),
		FourTimesADayMachine: getFourTimesADayMachine(
			params.FourTimesPerDayWordSet,
			params.PostSixWordSet,
			params.HourlySuffixSet),
		EveryOtherHourMachine: getEveryOtherHourMachine(params.EveryOtherHourWordSet),
		EveryOtherDayMachine: getEveryOtherDayMachine(
			params.EveryOtherDayWordSet,
			params.DailyWordSet),
		TwiceADayMachine: getTwiceADayMachine(
			params.TwiceADayWordSet,
			params.HourlySuffixSet),
		DailySuffixMachine:   getDailySuffixMachine(params.DailySuffixSet),
		WeeklyMachine:        getWeeklyMachine(params.PerWeekWordSet),
		HourlySuffixMachine:  getHourlySuffixMachine(params.HourlySuffixSet),
		WeeklySuffixMachine:  getWeeklySuffixMachine(params.WeeklySuffixSet),
		MonthlySuffixMachine: getMonthlySuffixMachine(params.MonthlySuffixSet),
		YearlySuffixMachine:  getYearlySuffixMachine(params.YearlySuffixSet),
		PrnMachine:           getAsNeededMachine(params.PrnWordSet),
	}

	return func(sent *types.Sentence, overrideSet []types.HasSpan) types.Spans {
		outSet := make(types.Spans, 0)

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

				if currentState == endState || currentState == ntEndState || currentState == skipFirstState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					startToken := tokens[tokenStartIndex]
					if currentState == skipFirstState {
						startToken = tokens[tokenStartIndex+1]
					}

					endToken := tokenSpan
					if currentState == ntEndState {
						endToken = tokens[i-1]
					}

					quantity := QuantityPrn
					switch machineIdx {
					//case DailyMachine:
					//	quantity = QuantityOne
					case SixTimesADayMachine:
						quantity = QuantitySix
					case FiveTimesADayMachine:
						quantity = QuantityFive
					case ThreeTimesADayMachine:
						quantity = QuantityThree
					case FourTimesADayMachine:
						quantity = QuantityFour
					//case EveryOtherHourMachine:
					//	quantity = Quantity24 / 2
					//case EveryOtherDayMachine:
					//	quantity = QuantityEveryOtherDay
					case TwiceADayMachine:
						quantity = QuantityTwo
						//case DailySuffixMachine:
						//	quantity = QuantityOne
						//case WeeklyMachine:
						//	quantity = QuantityWeekly
						//case HourlySuffixMachine:
						//	quantity = Quantity24
						//case WeeklySuffixMachine:
						//	quantity = QuantityWeekly
						//case MonthlySuffixMachine:
						//	quantity = QuantityMonthly
						//case YearlySuffixMachine:
						//	quantity = QuantityYearly
						//case PrnMachine:
						//	quantity = QuantityPrn
					}

					outToken := FrequencyUnitToken{
						Span: types.Span{
							Begin: startToken.GetSpan().Begin,
							End:   endToken.GetSpan().End,
						},
						Quantity: quantity,
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

func getWeeklyMachine(perWeekWordSet map[string]bool) fsm.Machine {

	soloCondition := fsm.NewWordSetCondition(perWeekWordSet)
	contCondition := fsm.NewContainsSetTextValueCondition(perWeekWordSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	qCondition := fsm.NewTextValueCondition("q")
	aCondition := fsm.NewTextValueCondition("a")
	atCondition := fsm.NewTextValueCondition("at")
	perCondition := fsm.NewTextValueCondition("per")
	wCondition := fsm.NewTextValueCondition("w")
	kCondition := fsm.NewTextValueCondition("k")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: qCondition, Dst: leftAbbreviateQState},
			{Cond: aCondition, Dst: leftAbbreviateState},
			{Cond: atCondition, Dst: leftAbbreviateState},
			{Cond: perCondition, Dst: leftAbbreviateState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateQState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotQState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: contCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: wCondition, Dst: middleAbbreviateQtoWState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoWState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoWState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoWState: []fsm.MachineRule{
			{Cond: kCondition, Dst: rightAbbreviateQWKState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQWKState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getThreeTimesADayMachine(
	threeTimesADayWordSet map[string]bool,
	postEightWordSet map[string]bool,
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	TTDCondition := fsm.NewWordSetCondition(threeTimesADayWordSet)
	postEightWordSetCondition := fsm.NewWordSetCondition(postEightWordSet)
	hourlySuffixSetCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')
	tCondition := fsm.NewTextValueCondition("t")
	qCondition := fsm.NewTextValueCondition("q")
	toCondition := fsm.NewTextValueCondition("to")
	tenCondition := fsm.NewTextValueCondition("ten")
	nineCondition := fsm.NewTextValueCondition("nine")
	iCondition := fsm.NewTextValueCondition("i")
	eightCondition := fsm.NewTextValueCondition("eight")
	dCondition := fsm.NewTextValueCondition("d")
	cond8 := NewIntegerValueCondition(8)

	numRange9to10 := fsm.NewIntegerRangeCondition(9, 10)

	combCondition := fsm.NewCombineCondition(
		fsm.NewNegateCondition(fsm.NewIntegerRangeCondition(1, 7)),
		fsm.NumberCondition,
	)
	disCond1 := fsm.NewDisjointCondition(combCondition, NewIntegerValueCondition(8))
	disCond2 := fsm.NewDisjointCondition(postEightWordSetCondition, eightCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: tCondition, Dst: leftAbbreviateTState},
			{Cond: qCondition, Dst: eightHourState},
			{Cond: cond8, Dst: eightHourState},
			{Cond: TTDCondition, Dst: endState},
			{Cond: disCond1, Dst: handleRangeState},
			{Cond: disCond2, Dst: eightHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		handleRangeState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		eightHourState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphState},
			{Cond: toCondition, Dst: hyphState},
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateTState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotTState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphState: []fsm.MachineRule{
			{Cond: numRange9to10, Dst: rangeState},
			{Cond: tenCondition, Dst: rangeState},
			{Cond: nineCondition, Dst: rangeState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotTState: []fsm.MachineRule{
			{Cond: iCondition, Dst: middleAbbreviateTtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rangeState: []fsm.MachineRule{
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateTtoIState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotTtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotTtoIState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateTIDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateTIDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getTwiceADayMachine(
	twiceADayWordSet map[string]bool,
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	TDCondition := fsm.NewWordSetCondition(twiceADayWordSet)
	hourlySuffixSetCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	bCondition := fsm.NewTextValueCondition("b")
	qCondition := fsm.NewTextValueCondition("q")
	twelveCondition := fsm.NewTextValueCondition("twelve")
	iCondition := fsm.NewTextValueCondition("i")
	dCondition := fsm.NewTextValueCondition("d")
	cond12 := NewIntegerValueCondition(12)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: bCondition, Dst: leftAbbreviateBState},
			{Cond: qCondition, Dst: twelveHourState},
			{Cond: cond12, Dst: twelveHourState},
			{Cond: twelveCondition, Dst: twelveHourState},
			{Cond: TDCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		twelveHourState: []fsm.MachineRule{
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: TDCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateBState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotBState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotBState: []fsm.MachineRule{
			{Cond: iCondition, Dst: middleAbbreviateBtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateBtoIState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotBtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotBtoIState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateBIDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateBIDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getEveryOtherDayMachine(
	everyOtherDayWordSet map[string]bool,
	dailyWordSet map[string]bool,
) fsm.Machine {

	EODCondition := fsm.NewWordSetCondition(everyOtherDayWordSet)
	dailyWordSetCondition := fsm.NewWordSetCondition(dailyWordSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	qCondition := fsm.NewTextValueCondition("q")
	everyOtherCondition := fsm.NewTextValueCondition("every-other")
	aCondition := fsm.NewTextValueCondition("a")
	dCondition := fsm.NewTextValueCondition("d")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: qCondition, Dst: leftAbbreviateQState},
			{Cond: EODCondition, Dst: endState},
			{Cond: everyOtherCondition, Dst: EODState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateQState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotQState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: aCondition, Dst: middleAbbreviateQtoAState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoAState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoAState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoAState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateQADState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		EODState: []fsm.MachineRule{
			{Cond: dailyWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQADState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getEveryOtherHourMachine(everyOtherHourWordSet map[string]bool) fsm.Machine {

	everyOtherHourWordSetCondition := fsm.NewWordSetCondition(everyOtherHourWordSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	qCondition := fsm.NewTextValueCondition("q")
	oCondition := fsm.NewTextValueCondition("o")
	dCondition := fsm.NewTextValueCondition("d")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: qCondition, Dst: leftAbbreviateQState},
			{Cond: everyOtherHourWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateQState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotQState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: oCondition, Dst: middleAbbreviateQtoOState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoOState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoOState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoOState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateQODState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQODState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getAsNeededMachine(prnWordSet map[string]bool) fsm.Machine {

	prnWordSetCondition := fsm.NewWordSetCondition(prnWordSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')
	pCondition := fsm.NewTextValueCondition("p")
	rCondition := fsm.NewTextValueCondition("r")
	nCondition := fsm.NewTextValueCondition("n")

	asCondition := fsm.NewTextValueCondition("as")
	neededCondition := fsm.NewTextValueCondition("needed")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: asCondition, Dst: asNeededState},
			{Cond: prnWordSetCondition, Dst: endState},
			{Cond: pCondition, Dst: startPState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startPState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: startPDOTState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startPDOTState: []fsm.MachineRule{
			{Cond: rCondition, Dst: startRState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startRState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: startRDOTState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startRDOTState: []fsm.MachineRule{
			{Cond: nCondition, Dst: startNState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		startNState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		asNeededState: []fsm.MachineRule{
			{Cond: neededCondition, Dst: endState},
			{Cond: dashCondition, Dst: asNeededHyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		asNeededHyphState: []fsm.MachineRule{
			{Cond: neededCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getFourTimesADayMachine(
	fourTimesPerDayWordSet map[string]bool,
	postSixWordSet map[string]bool,
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	fourTimesPerDayWordSetCondition := fsm.NewWordSetCondition(fourTimesPerDayWordSet)
	postSixWordSetCondition := fsm.NewWordSetCondition(postSixWordSet)
	hourlySuffixSetCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')
	dCondition := fsm.NewTextValueCondition("d")
	iCondition := fsm.NewTextValueCondition("i")
	sCondition := fsm.NewTextValueCondition("s")
	qCondition := fsm.NewTextValueCondition("q")

	toCondition := fsm.NewTextValueCondition("to")
	sixCondition := fsm.NewTextValueCondition("six")
	sevenCondition := fsm.NewTextValueCondition("seven")
	eightCondition := fsm.NewTextValueCondition("eight")
	nineCondition := fsm.NewTextValueCondition("nine")
	tenCondition := fsm.NewTextValueCondition("ten")

	int7to10Condition := fsm.NewIntegerRangeCondition(7, 10)
	int6Condition := NewIntegerValueCondition(6)

	combCondition := fsm.NewCombineCondition(fsm.NewNegateCondition(fsm.NewIntegerRangeCondition(1, 5)), fsm.NumberCondition)
	disjCondition1 := fsm.NewDisjointCondition(combCondition, NewIntegerValueCondition(6))
	disjCondition2 := fsm.NewDisjointCondition(postSixWordSetCondition, sixCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: qCondition, Dst: leftAbbreviateQState},
			{Cond: fourTimesPerDayWordSetCondition, Dst: endState},
			{Cond: int6Condition, Dst: sixHourState},
			{Cond: disjCondition1, Dst: handleRangeState},
			{Cond: disjCondition2, Dst: sixHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		handleRangeState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: rangeHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateQState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotQState},
			{Cond: sixCondition, Dst: sixHourState},
			{Cond: int6Condition, Dst: sixHourState},
			{Cond: fourTimesPerDayWordSetCondition, Dst: sixHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rangeHourState: []fsm.MachineRule{
			{Cond: int7to10Condition, Dst: eightSuffixState},
			{Cond: sevenCondition, Dst: eightSuffixState},
			{Cond: eightCondition, Dst: eightSuffixState},
			{Cond: nineCondition, Dst: eightSuffixState},
			{Cond: tenCondition, Dst: eightSuffixState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		eightSuffixState: []fsm.MachineRule{
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		sixHourState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: rangeHourState},
			{Cond: toCondition, Dst: rangeHourState},
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fourTimesPerDayWordSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: dCondition, Dst: middleAbbreviateQtoDState},
			{Cond: iCondition, Dst: middleAbbreviateQtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoDState: []fsm.MachineRule{
			{Cond: sCondition, Dst: rightAbbreviateQDSState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoIState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateQIDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoIState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoIState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQDSState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQIDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getSixTimesADayMachine(
	sixTimesPerDayWordSet map[string]bool,
	postFourWordSet map[string]bool,
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	sixTimesPerDayWordSetCondition := fsm.NewWordSetCondition(sixTimesPerDayWordSet)
	postFourWordSetCondition := fsm.NewWordSetCondition(postFourWordSet)
	hourlySuffixSetCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	dashCondition := fsm.NewPunctuationValueCondition('-')
	dCondition := fsm.NewTextValueCondition("d")
	sCondition := fsm.NewTextValueCondition("s")

	toCondition := fsm.NewTextValueCondition("to")

	fourCondition := fsm.NewTextValueCondition("four")
	fiveCondition := fsm.NewTextValueCondition("five")
	sixCondition := fsm.NewTextValueCondition("six")
	sevenCondition := fsm.NewTextValueCondition("seven")
	eightCondition := fsm.NewTextValueCondition("eight")

	int5to8Condition := fsm.NewIntegerRangeCondition(5, 8)
	int1to3Condition := fsm.NewIntegerRangeCondition(1, 3)
	int4Condition := NewIntegerValueCondition(4)

	combCondition := fsm.NewCombineCondition(fsm.NewNegateCondition(int1to3Condition), fsm.NumberCondition)
	disjCondition1 := fsm.NewDisjointCondition(combCondition, NewIntegerValueCondition(4))
	disjCondition2 := fsm.NewDisjointCondition(postFourWordSetCondition, fourCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: sixTimesPerDayWordSetCondition, Dst: endState},
			{Cond: int4Condition, Dst: fourHourState},
			{Cond: disjCondition1, Dst: handleRangeState},
			{Cond: disjCondition2, Dst: fourHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		handleRangeState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		fourHourState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphState},
			{Cond: toCondition, Dst: hyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: dCondition, Dst: middleAbbreviateQtoDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphState: []fsm.MachineRule{
			{Cond: int5to8Condition, Dst: numState},
			{Cond: fiveCondition, Dst: numState},
			{Cond: sixCondition, Dst: numState},
			{Cond: sevenCondition, Dst: numState},
			{Cond: eightCondition, Dst: numState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		numState: []fsm.MachineRule{
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoDState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoDState: []fsm.MachineRule{
			{Cond: sCondition, Dst: rightAbbreviateQDSState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQDSState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getDailyMachine(
	dailyWordSet map[string]bool,
	perDayWordSet map[string]bool,
) fsm.Machine {

	specificWordCondition := fsm.NewWordSetCondition(dailyWordSet)
	soloCondition := fsm.NewWordSetCondition(perDayWordSet)
	containsSoloTermCondition := fsm.NewContainsSetTextValueCondition(perDayWordSet)

	dotCondition := fsm.NewPunctuationValueCondition('.')
	dCondition := fsm.NewTextValueCondition("d")
	mCondition := fsm.NewTextValueCondition("m")
	sCondition := fsm.NewTextValueCondition("s")
	qCondition := fsm.NewTextValueCondition("q")
	oCondition := fsm.NewTextValueCondition("o")
	hCondition := fsm.NewTextValueCondition("h")
	aCondition := fsm.NewTextValueCondition("a")
	pCondition := fsm.NewTextValueCondition("p")

	bedCondition := fsm.NewTextValueCondition("bed")
	perCondition := fsm.NewTextValueCondition("per")
	timeCondition := fsm.NewTextValueCondition("time")

	int1to12Condition := fsm.NewIntegerRangeCondition(1, 12)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: qCondition, Dst: leftAbbreviateQState},
			{Cond: oCondition, Dst: leftAbbreviateOState},
			{Cond: hCondition, Dst: leftAbbreviateHState},
			{Cond: int1to12Condition, Dst: clockState},
			{Cond: TimeCondition, Dst: endState},
			{Cond: bedCondition, Dst: leftAbbreviateState},
			{Cond: perCondition, Dst: leftAbbreviateState},
			{Cond: specificWordCondition, Dst: endState},
			{Cond: soloCondition, Dst: endState},
			{Cond: containsSoloTermCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		clockState: []fsm.MachineRule{
			{Cond: aCondition, Dst: leftAbbreviateAState},
			{Cond: pCondition, Dst: leftAbbreviatePState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: specificWordCondition, Dst: endState},
			{Cond: timeCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateQState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotQState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateOState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotOState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateHState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotHState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotQState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateQDState},
			{Cond: hCondition, Dst: middleAbbreviateQtoHState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotOState: []fsm.MachineRule{
			{Cond: dCondition, Dst: rightAbbreviateODState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleAbbreviateQtoHState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: secondDotQtoHState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateAState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotAState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviatePState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotPState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotAState: []fsm.MachineRule{
			{Cond: mCondition, Dst: rightAbbreviateAMState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotPState: []fsm.MachineRule{
			{Cond: mCondition, Dst: rightAbbreviatePMState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDotQtoHState: []fsm.MachineRule{
			{Cond: sCondition, Dst: rightAbbreviateQHSState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateAMState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviatePMState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotHState: []fsm.MachineRule{
			{Cond: sCondition, Dst: rightAbbreviateHSState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateODState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQDState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateHSState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviateQHSState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getYearlySuffixMachine(
	yearlySuffixSet map[string]bool,
) fsm.Machine {

	suffixCondition := fsm.NewWordSetCondition(yearlySuffixSet)

	forwardSlashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: forwardSlashCondition, Dst: forwardSlashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		forwardSlashState: []fsm.MachineRule{
			{Cond: suffixCondition, Dst: skipFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		skipFirstState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getMonthlySuffixMachine(
	monthlySuffixSet map[string]bool,
) fsm.Machine {

	suffixCondition := fsm.NewWordSetCondition(monthlySuffixSet)

	forwardSlashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: forwardSlashCondition, Dst: forwardSlashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		forwardSlashState: []fsm.MachineRule{
			{Cond: suffixCondition, Dst: skipFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		skipFirstState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getWeeklySuffixMachine(
	weeklySuffixSet map[string]bool,
) fsm.Machine {

	suffixCondition := fsm.NewWordSetCondition(weeklySuffixSet)

	forwardSlashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: forwardSlashCondition, Dst: forwardSlashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		forwardSlashState: []fsm.MachineRule{
			{Cond: suffixCondition, Dst: skipFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		skipFirstState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getDailySuffixMachine(
	dailySuffixSet map[string]bool,
) fsm.Machine {

	suffixCondition := fsm.NewWordSetCondition(dailySuffixSet)

	forwardSlashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: forwardSlashCondition, Dst: forwardSlashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		forwardSlashState: []fsm.MachineRule{
			{Cond: suffixCondition, Dst: skipFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		skipFirstState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getHourlySuffixMachine(
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	suffixCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	forwardSlashCondition := fsm.NewPunctuationValueCondition('/')

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: forwardSlashCondition, Dst: forwardSlashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		forwardSlashState: []fsm.MachineRule{
			{Cond: suffixCondition, Dst: skipFirstState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		skipFirstState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getFiveTimesADayMachine(
	sixTimesPerDayWordSet map[string]bool,
	hourlySuffixSet map[string]bool,
) fsm.Machine {

	sixTimesPerDayWordSetCondition := fsm.NewWordSetCondition(sixTimesPerDayWordSet)
	hourlySuffixSetCondition := fsm.NewWordSetCondition(hourlySuffixSet)

	dashCondition := fsm.NewPunctuationValueCondition('-')
	fiveCondition := fsm.NewTextValueCondition("five")

	int1to4Condition := fsm.NewIntegerRangeCondition(1, 4)
	int5to10Condition := fsm.NewIntegerRangeCondition(5, 10)
	int5Condition := NewIntegerValueCondition(5)

	combCondition := fsm.NewCombineCondition(fsm.NewNegateCondition(int1to4Condition), int5Condition)
	disjCondition := fsm.NewDisjointCondition(combCondition, NewIntegerValueCondition(4))

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: sixTimesPerDayWordSetCondition, Dst: endState},
			{Cond: disjCondition, Dst: handleRangeState},
			{Cond: int5Condition, Dst: fiveHourState},
			{Cond: fiveCondition, Dst: fiveHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		handleRangeState: []fsm.MachineRule{
			{Cond: dashCondition, Dst: hyphState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		fiveHourState: []fsm.MachineRule{
			{Cond: hourlySuffixSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		hyphState: []fsm.MachineRule{
			{Cond: int5to10Condition, Dst: fiveHourState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntEndState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
