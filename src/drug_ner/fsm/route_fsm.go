package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
)

type RouteFSMParams struct {
	SpecifiedOralWordSet    map[string]bool `drug:"specified_oral_word_set.txt"`
	SpecifiedPatchesWordSet map[string]bool `drug:"specified_patches_word_set.txt"`
	SpecifiedGastricWordSet map[string]bool `drug:"specified_gastric_word_set.txt"`
	SingleTopicalWordSet    map[string]bool `drug:"single_topical_word_set.txt"`
	SingleOralWordSet       map[string]bool `drug:"single_oral_word_set.txt"`
	SingleRectalWordSet     map[string]bool `drug:"single_rectal_word_set.txt"`
	SingleInjectWordSet     map[string]bool `drug:"single_inject_word_set.txt"`
	MiddleTermSet           map[string]bool `drug:"middle_term_set.txt"`
}

type FormMethod string

const (
	Topical     FormMethod = "Topical"
	Oral        FormMethod = "Enteral_Oral"
	Gastric     FormMethod = "Enteral_Gastric"
	Rectal      FormMethod = "Enteral_Rectal"
	Intravenous FormMethod = "Parenteral_Intravenous"
	Transdermal FormMethod = "Parenteral_Transdermal"
)

type RouteToken struct {
	types.Span
	FormMethod FormMethod
}

func (token RouteToken) GetSpan() *types.Span {
	return &token.Span
}

type RouteFSM func(sent *types.Sentence) types.Spans

func NewRouteFSM(params RouteFSMParams) RouteFSM {

	patchesMachineIndex := 0
	gastricMachineIndex := 1
	topicalMachineIndex := 2
	oralMachineIndex := 3
	rectalMachineIndex := 4
	injectMachineIndex := 5

	machineSet := map[int]fsm.Machine{
		patchesMachineIndex: getPatchesMachine(params.MiddleTermSet, params.SpecifiedPatchesWordSet),
		gastricMachineIndex: getGastricMachine(params.MiddleTermSet, params.SpecifiedGastricWordSet),
		topicalMachineIndex: getTopicalMachine(params.MiddleTermSet, params.SingleTopicalWordSet),
		oralMachineIndex:    getOralMachine(params.MiddleTermSet, params.SingleOralWordSet, params.SpecifiedOralWordSet),
		rectalMachineIndex:  getRectalMachine(params.SingleRectalWordSet),
		injectMachineIndex:  getInjectionMachine(params.SingleInjectWordSet),
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
			var tokenSpan types.HasSpan = token

			for machineIdx, machine := range machineSet {
				currentState := machineStates[machineIdx]
				currentState = machine.Input(tokenSpan, currentState)
				machineStates[machineIdx] = currentState

				if currentState == startState {
					tokenStartMap[machineIdx] = i
				}

				if currentState == endState {
					tokenStartIndex, isOk := tokenStartMap[machineIdx]
					if isOk {
						tokenStartIndex++
					}

					frmMethod := Intravenous
					switch machineIdx {
					case patchesMachineIndex:
						frmMethod = Transdermal
					case gastricMachineIndex:
						frmMethod = Gastric
					case topicalMachineIndex:
						frmMethod = Topical
					case oralMachineIndex:
						frmMethod = Oral
					case rectalMachineIndex:
						frmMethod = Rectal
					}

					outToken := RouteToken{
						Span: types.Span{
							Begin: tokens[tokenStartIndex].GetSpan().Begin,
							End:   tokenSpan.GetSpan().End,
						},
						FormMethod: frmMethod,
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

func getPatchesMachine(
	middleTermSet map[string]bool,
	specifiedPatchesWordSet map[string]bool,
) fsm.Machine {

	middleTermCondition := fsm.NewWordSetCondition(middleTermSet)
	specificWordCondition := fsm.NewWordSetCondition(specifiedPatchesWordSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)
	aCondition := fsm.NewTextValueCondition("a")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: aCondition, Dst: leftAbbreviateState},
			{Cond: middleTermCondition, Dst: leftAbbreviateState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: specificWordCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: dashCondition, Dst: secondDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
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

func getGastricMachine(
	middleTermSet map[string]bool,
	specifiedGastricWordSet map[string]bool,
) fsm.Machine {

	middleTermCondition := fsm.NewWordSetCondition(middleTermSet)
	specificWordCondition := fsm.NewWordSetCondition(specifiedGastricWordSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)
	aCondition := fsm.NewTextValueCondition("a")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: aCondition, Dst: leftAbbreviateState},
			{Cond: middleTermCondition, Dst: leftAbbreviateState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: specificWordCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: dashCondition, Dst: secondDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
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

func getTopicalMachine(
	middleTermSet map[string]bool,
	singleTopicalWordSet map[string]bool,
) fsm.Machine {

	middleTermCondition := fsm.NewWordSetCondition(middleTermSet)
	soloCondition := fsm.NewWordSetCondition(singleTopicalWordSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')
	dotCondition := fsm.NewPunctuationValueCondition('.')
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)
	aCondition := fsm.NewTextValueCondition("a")
	pCondition := fsm.NewTextValueCondition("p")
	vCondition := fsm.NewTextValueCondition("v")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: aCondition, Dst: leftAbbreviateState},
			{Cond: pCondition, Dst: leftAbbreviatePState},
			{Cond: middleTermCondition, Dst: leftAbbreviateState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviatePState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotPState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotPState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: vCondition, Dst: rightAbbreviatePVState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: dashCondition, Dst: secondDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviatePVState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getOralMachine(
	middleTermSet map[string]bool,
	singleOralWordSet map[string]bool,
	specifiedOralWordSet map[string]bool,
) fsm.Machine {

	middleTermCondition := fsm.NewWordSetCondition(middleTermSet)
	soloCondition := fsm.NewWordSetCondition(singleOralWordSet)
	specificWordCondition := fsm.NewWordSetCondition(specifiedOralWordSet)
	dashCondition := fsm.NewPunctuationValueCondition('-')
	dotCondition := fsm.NewPunctuationValueCondition('.')
	rangeCombineCondition := fsm.NewDisjointCondition(RangeCondition, RangeStrengthCondition)
	aCondition := fsm.NewTextValueCondition("a")
	pCondition := fsm.NewTextValueCondition("p")
	oCondition := fsm.NewTextValueCondition("o")

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: aCondition, Dst: leftAbbreviateState},
			{Cond: pCondition, Dst: leftAbbreviatePState},
			{Cond: middleTermCondition, Dst: leftAbbreviateState},
			{Cond: rangeCombineCondition, Dst: leftDosagesState},
			{Cond: soloCondition, Dst: endState},
			{Cond: specificWordCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviatePState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: firstDotPState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDotPState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: oCondition, Dst: rightAbbreviatePOState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		leftAbbreviateState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: dashCondition, Dst: firstDashState},
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		firstDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: middleATextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		middleATextState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: dashCondition, Dst: secondDashState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		secondDashState: []fsm.MachineRule{
			{Cond: middleTermCondition, Dst: lastTextState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		lastTextState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		rightAbbreviatePOState: []fsm.MachineRule{
			{Cond: dotCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getRectalMachine(
	singleRectalWordSet map[string]bool,
) fsm.Machine {

	soloCondition := fsm.NewWordSetCondition(singleRectalWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}

func getInjectionMachine(
	singleInjectWordSet map[string]bool,
) fsm.Machine {

	soloCondition := fsm.NewWordSetCondition(singleInjectWordSet)

	return fsm.Machine{
		startState: []fsm.MachineRule{
			{Cond: soloCondition, Dst: endState},
			{Cond: fsm.AnyCondition, Dst: startState},
		},
		endState: []fsm.MachineRule{
			{Cond: fsm.AnyCondition, Dst: startState},
		},
	}
}
