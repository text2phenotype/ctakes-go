package pos

import "text2phenotype.com/fdl/types"

type SequenceValidator interface {
	ValidSequence(i int, inputSequence []*types.Token, outcome string) bool
}

type defaultSequenceValidator struct {
	tagDictionary map[string]map[string]bool
}

func (g defaultSequenceValidator) ValidSequence(i int, inputSequence []*types.Token, outcome string) bool {
	if g.tagDictionary != nil {
		tags, res := g.tagDictionary[*inputSequence[i].Text]
		if !res {
			return true
		}

		return tags[outcome]
	}

	return true
}

func NewSequenceValidator() SequenceValidator {
	return defaultSequenceValidator{}
}
