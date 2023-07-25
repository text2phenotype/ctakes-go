package fsm

import (
	"bytes"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"io/ioutil"
	"sort"
	"strings"
	"unicode"
)

type LabUnitsFSM interface {
	Execute(tokens []types.HasSpan) types.HasSpan
	Split(tokens []types.Token) []types.Token
}

type labUnitsFSM struct {
	unitTree *utils.PrefixTree
}

func (fsm labUnitsFSM) Execute(tokens []types.HasSpan) types.HasSpan {
	var unitToken types.Token

	currentNode := fsm.unitTree.Root
	for _, token := range tokens {
		tokenSpan := token.GetSpan()
		for _, c := range *tokenSpan.Text {
			childNode, isOk := currentNode.Children[c]
			if !isOk {
				if unitToken.Begin != 0 {
					return &unitToken
				}

				return nil
			}

			currentNode = childNode
		}

		if len(currentNode.Text) > 0 {
			unitToken.Text = &currentNode.Text
			unitToken.End = tokenSpan.End
			unitToken.Begin = tokens[0].GetSpan().Begin
		}
	}

	if unitToken.Begin != 0 {
		return &unitToken
	}

	return nil
}

/*
Splits tokens if it contains value and unit without separator
e.g "45mg/dL" -> "45" and "mg/dL"
*/
func (fsm labUnitsFSM) Split(tokens []types.Token) []types.Token {
	var result []types.Token

	for _, token := range tokens {
		newToken := token.Clone()
		if len(*token.Text) >= 2 {

			currentNode := fsm.unitTree.Root
			tokenSpan := token.GetSpan()
			var unitLen int32

			runes := []rune(*tokenSpan.Text)

			if unicode.IsDigit(runes[0]) {
				for rIdx, c := range runes {
					if currentNode == fsm.unitTree.Root && unicode.IsDigit(c) {
						continue
					}

					if rIdx > 0 && !unicode.IsDigit(c) {
						childNode, isOk := currentNode.Children[c]
						if !isOk {
							currentNode = fsm.unitTree.Root
							unitLen = 0
							break
						}

						unitLen++
						currentNode = childNode
					}

				}

				newtokenLen := int32(len(*newToken.Text)) - unitLen
				if currentNode != fsm.unitTree.Root && newtokenLen > 0 {

					shapedText := []rune(newToken.GetShapedText())

					unitText := string(shapedText[newtokenLen:])
					lowerText := strings.ToLower(unitText)

					unitToken := types.Token{
						Span: types.Span{
							Begin: newToken.End - unitLen,
							End:   newToken.End,
							Text:  &lowerText,
						},
						Shape:  types.GetShape(unitText),
						IsWord: true,
					}
					result = append(result, unitToken)

					newToken.End = newToken.Begin + newtokenLen
					newTokenTxt := string(shapedText[:newtokenLen])
					newTokenTxtLower := strings.ToLower(newTokenTxt)
					newToken.Text = &newTokenTxtLower
					newToken.Shape = types.GetShape(newTokenTxt)
					newToken.IsNumber = true
					newToken.IsWord = false

				}
			}

		}
		result = append(result, newToken)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Begin < result[j].Begin
	})

	return result
}

func buildUnitsTree(units []string) *utils.PrefixTree {
	tree := utils.PrefixTree{
		Root: &utils.PrefixTreeNode{
			Children: make(map[rune]*utils.PrefixTreeNode),
		},
	}

	for _, unit := range units {
		node := tree.Root
		unit := strings.ToLower(strings.Trim(unit, "\n"))
		for _, c := range unit {
			childNode, isOk := node.Children[c]
			if !isOk {
				childNode = &utils.PrefixTreeNode{
					Parent:   node,
					Children: make(map[rune]*utils.PrefixTreeNode),
				}
			}

			node.Children[c] = childNode
			node = childNode
		}
		node.Text = unit
	}

	return &tree
}

func NewLabUnitsFSM(unitsFilePath string) (LabUnitsFSM, error) {
	b, err := ioutil.ReadFile(unitsFilePath)
	if err != nil {
		return nil, err
	}

	var units []string
	buf := bytes.NewBuffer(b)
	line, err := buf.ReadString('\n')
	for err == nil {
		units = append(units, line)
		line, err = buf.ReadString('\n')
	}

	return &labUnitsFSM{
		unitTree: buildUnitsTree(units),
	}, nil
}
