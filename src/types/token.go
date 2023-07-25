package types

import (
	"strings"
	"unicode"
)

type Token struct {
	Span
	Tag       *string
	Sentence  *Sentence
	Lemma     *string
	IsPunct   bool
	IsWord    bool
	IsSymbol  bool
	IsNumber  bool
	IsNewline bool
	Shape     string
}

func (token *Token) GetSpan() *Span {
	return &token.Span
}

func (token *Token) GetShapedText() string {
	var sb strings.Builder
	runes := []rune(*token.Text)
	if len(runes) > len(token.Shape) {
		return *token.Text
	}
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		if token.Shape[i] == 'X' {
			sb.WriteRune(unicode.ToUpper(ch))
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

func (token Token) Clone() Token {
	return Token{
		Span: Span{
			Begin: token.Begin,
			End:   token.End,
			Text:  token.Text,
		},
		Tag:       token.Tag,
		Sentence:  token.Sentence,
		Lemma:     token.Lemma,
		IsPunct:   token.IsPunct,
		IsWord:    token.IsWord,
		IsSymbol:  token.IsSymbol,
		IsNumber:  token.IsNumber,
		IsNewline: token.IsNewline,
		Shape:     token.Shape,
	}
}

func GetShape(txt string) string {
	var sb strings.Builder
	for _, r := range txt {
		switch {
		case unicode.IsDigit(r):
			sb.WriteRune('d')
		case unicode.IsUpper(r):
			sb.WriteRune('X')
		default:
			sb.WriteRune('x')
		}
	}

	return sb.String()
}
