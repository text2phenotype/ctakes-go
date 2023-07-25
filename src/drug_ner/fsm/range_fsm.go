package fsm

import "text2phenotype.com/fdl/types"

type RangeToken struct {
	types.Span
}

func (token RangeToken) GetSpan() *types.Span {
	return &token.Span
}
