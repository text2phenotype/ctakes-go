package lab

import "text2phenotype.com/fdl/types"

type ValueToken struct {
	token *types.Token
}

func (val *ValueToken) GetSpan() *types.Span {
	return val.token.GetSpan()
}
