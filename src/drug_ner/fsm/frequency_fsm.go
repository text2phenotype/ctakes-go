package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"errors"
	"strings"
	"unicode/utf8"
)

type FrequencyFSMParams struct {
	FrequencySet  map[string]string `drug:"frequency_set.txt"`
	MiddleTermSet map[string]bool   `drug:"middle_term_set.txt"`
	HyphenatedSet map[string]string `drug:"hyphenated_set.txt"`
}

type FrequencyToken struct {
	types.Span
	Value string
}

func (token FrequencyToken) GetSpan() *types.Span {
	return &token.Span
}

type FrequencyFSM func(sent *types.Sentence, overrideSet1 types.Spans, overrideSet2 types.Spans) types.Spans

func NewFrequencyFSM(params FrequencyFSMParams) FrequencyFSM {

	machineSet := []fsm.Machine{getFrequencyMachine(
		params.MiddleTermSet,
		params.FrequencySet,
		params.HyphenatedSet)}

	return func(sent *types.Sentence, overrideSet1 types.Spans, overrideSet2 types.Spans) types.Spans {
		outSet := make(types.Spans, 0)

		tokenStartMap := make(map[int]int)

		overrideTokenMap1 := make(map[int32]types.HasSpan)
		overrideTokenMap2 := make(map[int32]types.HasSpan)
		overrideBeginTokenMap1 := make(map[int32]int32)
		overrideBeginTokenMap2 := make(map[int32]int32)
		for _, t := range overrideSet1 {
			key := t.GetSpan().Begin
			overrideTokenMap1[key] = t
		}

		for _, t := range overrideSet2 {
			key := t.GetSpan().Begin
			overrideTokenMap2[key] = t
		}

		// init states
		machineStates := make([]string, len(machineSet))
		for n := 0; n < len(machineSet); n++ {
			machineStates[n] = startState
		}

		overrideOn1 := false
		overrideOn2 := false
		var overrideEndOffset1 int32 = -1
		var overrideEndOffset2 int32 = -1
		var tokenOffset1 int32 = 0
		var tokenOffset2 int32 = 0
		var anchorKey1 int32 = 0
		var anchorKey2 int32 = 0

		tokens := sent.Tokens
		for i, token := range tokens {
			var tokenSpan types.HasSpan = token
			key := tokenSpan.GetSpan().Begin

			if overrideOn1 && overrideOn2 {
				if overrideEndOffset1 >= overrideEndOffset2 {
					overrideOn1 = false
				} else {
					overrideOn2 = false
				}
			}

			if overrideOn1 {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset1 {
					overrideBeginTokenMap1[anchorKey1] = tokenOffset1
					overrideOn1 = false
					overrideEndOffset1 = -1
				} else {
					tokenOffset1++
					continue
				}
			} else if overrideOn2 {
				if tokenSpan.GetSpan().Begin >= overrideEndOffset2 {
					overrideBeginTokenMap2[anchorKey2] = tokenOffset2
					overrideOn2 = false
					overrideEndOffset2 = -1
				} else {
					tokenOffset2++
					continue
				}
			} else {

				if overToken, isOk := overrideTokenMap1[key]; isOk {
					anchorKey1 = key
					tokenSpan = overToken
					overrideOn1 = true
					overrideEndOffset1 = tokenSpan.GetSpan().End
					tokenOffset1 = 0
				}
				if overToken, isOk := overrideTokenMap2[key]; isOk {
					anchorKey2 = key
					tokenSpan = overToken
					overrideOn2 = true
					overrideEndOffset2 = tokenSpan.GetSpan().End
					tokenOffset2 = 0
				}
			}

			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
					tokenOffset1 = 0
					tokenOffset2 = 0
				}

				if currentState == endState || currentState == ntEndState || currentState == ntFalseTermState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						var tokenMap1 int32 = 0
						var tokenMap2 int32 = 0

						lookUpOffset := tokens[tokenStartIndex]

						if offSet, isOk2 := overrideBeginTokenMap1[lookUpOffset.GetSpan().Begin]; isOk2 {
							tokenMap1 = offSet + tokenMap1
						}
						if offSet, isOk2 := overrideBeginTokenMap2[lookUpOffset.GetSpan().Begin]; isOk2 {
							tokenMap2 = offSet + tokenMap2
						}

						globalOffset := tokenMap1 + tokenMap2
						tokenStartIndex += int(globalOffset)

						tokenStartIndex++
					}

					startToken := tokens[tokenStartIndex]
					if currentState == ntFalseTermState {
						startToken = tokens[tokenStartIndex+1]
					}

					endToken := tokenSpan
					if currentState == ntEndState {
						endToken = tokens[i-1]
					}

					outToken, err := createFrequencyToken(startToken, endToken, sent, params)
					if err != nil {
						continue
					}

					outSet = append(outSet, &outToken)
					machineStates[machineIdx] = startState
				}
			}
		}
		return outSet
	}
}

func createFrequencyToken(startToken, endToken types.HasSpan, sent *types.Sentence, params FrequencyFSMParams) (FrequencyToken, error) {
	outToken := FrequencyToken{
		Span: types.Span{
			Begin: startToken.GetSpan().Begin,
			End:   endToken.GetSpan().End,
		},
	}

	tokText, isOk := outToken.GetTextFromSentence(sent)
	if !isOk {
		return outToken, errors.New("frequency fsm: text is not found in the sentence")
	}

	lowerText := strings.ToLower(tokText)
	for kwrd, num := range params.FrequencySet {
		pos := strings.Index(lowerText, kwrd)
		if pos > -1 {
			outToken.Begin += int32(pos)
			outToken.End = outToken.Begin + int32(utf8.RuneCountInString(kwrd))
			outToken.Text = utils.GlobalStringStore().GetPointer(kwrd)
			outToken.Value = num
			return outToken, nil
		}

		pos = strings.Index(lowerText, num)
		if pos > -1 {
			outToken.Begin += int32(pos)
			outToken.End = outToken.Begin + int32(utf8.RuneCountInString(num))
			outToken.Text = utils.GlobalStringStore().GetPointer(num)
			outToken.Value = num
			return outToken, nil
		}
	}

	outToken.Text = utils.GlobalStringStore().GetPointer(lowerText)
	return outToken, nil
}

func getFrequencyMachine(
	middleTermSet map[string]bool,
	frequencySet map[string]string,
	hyphenatedSet map[string]string,
) fsm.Machine {

	middleTermCondition := fsm.NewWordSetCondition(middleTermSet)
	frequencySetCondition := fsm.NewWordMapCondition(frequencySet)
	hyphenatedSetCondition := fsm.NewWordMapCondition(hyphenatedSet)
	integerCondition := fsm.NewIntegerRangeCondition(0, 5)
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: rangeCombineCondition, Dst: leftAbbreviateState},
			{Cond: frequencySetCondition, Dst: leftAbbreviateState},
			{Cond: integerCondition, Dst: leftAbbreviateState},
			{Cond: hyphenatedSetCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: fsm.NumberCondition, Dst: middleATextState},
			{Cond: frequencySetCondition, Dst: midTermState},
			{Cond: hyphenatedSetCondition, Dst: endState},
			{Cond: FrequencyUnitCondition, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		midTermState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: termState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: FrequencyUnitCondition, Dst: ntEndState},
			{Cond: RouteCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		termState: []fsm.MachineRule{
			{Cond: FrequencyUnitCondition, Dst: ntFalseTermState},
			{Cond: RouteCondition, Dst: ntFalseTermState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: FrequencyUnitCondition, Dst: ntEndState},
			{Cond: RouteCondition, Dst: ntEndState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		ntFalseTermState: []fsm.MachineRule{
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
