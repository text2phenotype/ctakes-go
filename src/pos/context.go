package pos

import (
	"text2phenotype.com/fdl/types"
	"strings"
)

const (
	prefixLength = 4
	suffixLength = 4
)

type ContextGenerator interface {
	GetContext(index int, seq []*types.Token, priorDecisions []string) []string
}

type defaultContextGenerator struct {
	dict    map[string]bool
	seToken *types.Token
	sbToken *types.Token
}

func (g *defaultContextGenerator) GetContext(index int, tokens []*types.Token, tags []string) []string {
	var next, prev, nextnext, prevprev *types.Token
	var tagprev, tagprevprev string

	next = g.seToken
	prev = g.sbToken

	lex := tokens[index].GetShapedText()
	if len(tokens) > index+1 {
		next = tokens[index+1]
		nextnext = g.seToken
		if len(tokens) > index+2 {
			nextnext = tokens[index+2]
		}
	}

	if index > 0 {
		prev = tokens[index-1]
		prevprev = g.sbToken
		tagprev = tags[index-1]

		if index >= 2 {
			prevprev = tokens[index-2]
			tagprevprev = tags[index-2]
		}
	}

	var contexts []string
	contexts = append(contexts, "default", "w="+lex)

	if isOk := g.dict[lex]; !isOk {
		suffs := getSuffixes(lex)
		for _, suf := range suffs {
			contexts = append(contexts, "suf="+suf)
		}

		pefs := getPrefixes(lex)
		for _, pref := range pefs {
			contexts = append(contexts, "pre="+pref)
		}

		if strings.ContainsRune(lex, '-') {
			contexts = append(contexts, "h")
		}

		if strings.ContainsRune(tokens[index].Shape, 'X') {
			contexts = append(contexts, "c")
		}

		if strings.ContainsRune(tokens[index].Shape, 'd') {
			contexts = append(contexts, "d")
		}
	}

	contexts = append(contexts, "p="+prev.GetShapedText())

	if len(tagprev) > 0 {
		contexts = append(contexts, "t="+tagprev)
	}

	if prevprev != nil {
		contexts = append(contexts, "pp="+prevprev.GetShapedText())

		if len(tagprevprev) > 0 {
			contexts = append(contexts, "t2="+tagprevprev+","+tagprev)
		}
	}

	contexts = append(contexts, "n="+next.GetShapedText())
	if nextnext != nil {
		contexts = append(contexts, "nn="+nextnext.GetShapedText())
	}

	return contexts
}

func getPrefixes(lex string) []string {
	prefs := make([]string, prefixLength)
	for li := 0; li < prefixLength; li++ {
		idx := len(lex)
		if idx > li+1 {
			idx = li + 1
		}
		prefs[li] = lex[:idx]
	}
	return prefs
}

func getSuffixes(lex string) []string {
	suffs := make([]string, suffixLength)
	for li := 0; li < suffixLength; li++ {
		idx := len(lex) - li - 1
		if idx < 0 {
			idx = 0
		}

		suffs[li] = lex[idx:]
	}
	return suffs
}

func NewContextGenerator() ContextGenerator {
	se := "*SE*"
	sb := "*SB*"

	res := defaultContextGenerator{
		sbToken: &types.Token{Span: types.Span{Text: &sb}},
		seToken: &types.Token{Span: types.Span{Text: &se}},
	}

	return &res
}
