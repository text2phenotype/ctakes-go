package negation

import (
	"text2phenotype.com/fdl/types"
	"errors"
)

var (
	// errors
	SentenceIsNilError    error = errors.New("polarity analyzer: annotation sentence is nil")
	TokensNotFound        error = errors.New("polarity analyzer: sentence doesn't contain tokens")
	ScopeIsNotImplemented error = errors.New("polarity analyzer: scope is not implemented")
)

type PolarityAnalyzer func(annotations []types.Annotation, scopes []types.Scope) ([]types.Polarity, error)

func NewPolarityAnalyzer(maxLeftScopeSize int, maxRightScopeSize int, boundaries map[string]bool) PolarityAnalyzer {

	ctxAnalyzer := NewPolarityContextAnalyzer(boundaries)

	sizes := map[types.Scope]int{
		types.ScopeLeft:   maxLeftScopeSize,
		types.ScopeRight:  maxRightScopeSize,
		types.ScopeMiddle: 0,
	}

	return func(annotations []types.Annotation, scopes []types.Scope) ([]types.Polarity, error) {
		polarities := make([]types.Polarity, len(annotations))
		for i, annotation := range annotations {
			polarities[i] = types.PolarityPositive
			for _, scope := range scopes {
				scopeSize := sizes[scope]
				tokens, err := getScopeTokens(ctxAnalyzer, scopeSize, annotation, scope)
				if err != nil {
					return nil, err
				}

				if ctxAnalyzer.AnalyzeContext(tokens) {
					polarities[i] = types.PolarityNegative
				}
			}
		}
		return polarities, nil
	}

}

func getScopeTokens(ctxAnalyzer ContextAnalyzer, scopeSize int, annotation types.Annotation, scope types.Scope) ([]*types.Token, error) {
	sentence := annotation.Sentence
	if sentence == nil {
		return nil, SentenceIsNilError
	}

	tokens := sentence.Tokens
	if len(tokens) == 0 {
		return nil, TokensNotFound
	}

	startTokenIdx := -1
	endTokenIdx := -1

	for i := 0; i < len(tokens); i++ {
		if startTokenIdx < 0 && tokens[i].Begin >= annotation.Begin {
			startTokenIdx = i
		}

		if tokens[i].Begin >= annotation.Begin && tokens[i].End <= annotation.End {
			endTokenIdx = i
		}
	}

	scopeTokens := make([]*types.Token, 0, scopeSize+1)
	switch scope {
	case types.ScopeLeft:
		{
			for i := startTokenIdx - 1; i >= 0 && len(scopeTokens) < scopeSize; i-- {

				if ctxAnalyzer.IsBoundary(tokens[i]) {
					break
				}

				scopeTokens = append(scopeTokens, tokens[i])
			}

			if len(scopeTokens) > 1 {
				// reverse collection
				n := len(scopeTokens) / 2
				for i := 0; i < n; i++ {
					scopeTokens[i], scopeTokens[len(scopeTokens)-i-1] = scopeTokens[len(scopeTokens)-i-1], scopeTokens[i]
				}
			}
		}
	case types.ScopeRight:
		{
			for i := endTokenIdx + 1; i < len(tokens) && len(scopeTokens) < scopeSize; i++ {

				if ctxAnalyzer.IsBoundary(tokens[i]) {
					break
				}

				scopeTokens = append(scopeTokens, tokens[i])
			}
		}
	case types.ScopeMiddle:
		{
			return nil, ScopeIsNotImplemented
		}
	}

	scopeTokens = append(scopeTokens, getEOSToken())
	return scopeTokens, nil
}

func getEOSToken() *types.Token {
	txt := "<EOS>"
	return &types.Token{
		Span: types.Span{
			Text: &txt,
		},
	}
}

func NewPolarityContextAnalyzer(boundaries map[string]bool) ContextAnalyzer {
	var analyzer polarityContextAnalyzer
	analyzer.boundaryWordSet = boundaries
	analyzer.polarityFSM = NewPolarityFSM()
	return analyzer
}

func GetDefaultBoundaries() map[string]bool {
	return map[string]bool{
		"but":             true,
		"however":         true,
		"nevertheless":    true,
		"notwithstanding": true,
		"though":          true,
		"although":        true,
		"when":            true,
		"how":             true,
		"what":            true,
		"which":           true,
		"while":           true,
		"since":           true,
		"then":            true,
		"i":               true,
		"he":              true,
		"she":             true,
		"they":            true,
		"we":              true,
		";":               true,
		".":               true,
		")":               true,
	}
}

type polarityContextAnalyzer struct {
	boundaryWordSet map[string]bool
	polarityFSM     PolarityFSM
}

func (polarityCtxAnalyzer polarityContextAnalyzer) IsBoundary(token *types.Token) bool {
	txt := *token.Text
	return polarityCtxAnalyzer.boundaryWordSet[txt]
}

func (polarityCtxAnalyzer polarityContextAnalyzer) AnalyzeContext(tokens []*types.Token) bool {
	return polarityCtxAnalyzer.polarityFSM(tokens)
}
