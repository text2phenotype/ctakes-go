package pos

import "text2phenotype.com/fdl/types"

func NewTagger(model Model) func(tokens []*types.Token) []string {
	search := NewBeamSearch(model, 3)
	ctx := NewContextGenerator()
	validator := NewSequenceValidator()

	return func(tokens []*types.Token) []string {
		res, isOk := search(tokens, ctx, validator)
		if !isOk {
			return []string{}
		}
		return res.Outcomes
	}
}
