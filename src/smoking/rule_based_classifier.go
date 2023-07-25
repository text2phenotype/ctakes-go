package smoking

import (
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"strings"
)

type RuleBasedClassifierParams struct {
	SmokingWords map[string]bool `json:"smoking_words"`
	UnknownWords []string        `json:"unknown_words"`
}

type NominalAttributeValue struct {
}

func buildPrefixTree(phrases []string) utils.StringPrefixTree {

	var result utils.StringPrefixTree
	for _, phrase := range phrases {
		phrase = strings.TrimSpace(strings.ToLower(phrase))
		tokens := strings.Split(phrase, " ")
		result.Add(tokens, phrase)
	}
	return result
}

func NewRuleBasedClassifier(params RuleBasedClassifierParams) func(sent types.Sentence) string {
	unkWrdTree := buildPrefixTree(params.UnknownWords)

	return func(sent types.Sentence) string {
		classVal := ClassUnknown

		for _, token := range sent.Tokens {
			tokenText := strings.ToLower(*token.Text)
			isOk := params.SmokingWords[tokenText]
			if isOk {
				classVal = ClassKnown
				break
			}

		}

		if classVal != ClassKnown {
			return classVal
		}

		unkState := unkWrdTree.Root
		for _, token := range sent.Tokens {
			if chldNode, isOk := unkState.Children[*token.Text]; isOk {
				unkState = chldNode
			} else {
				unkState = unkWrdTree.Root
			}

			if len(unkState.Text) > 0 {
				classVal = ClassUnknown
				break
			}

		}
		return classVal
	}
}
