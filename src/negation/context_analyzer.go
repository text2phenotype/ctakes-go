package negation

import (
	"text2phenotype.com/fdl/types"
)

type ContextAnalyzer interface {
	IsBoundary(annotation *types.Token) bool
	AnalyzeContext(tokens []*types.Token) bool
}
