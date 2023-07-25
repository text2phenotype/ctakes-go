package types

type SentenceAttributes struct {
	SmokingStatus string
}

type Sentence struct {
	Span
	Tokens     []*Token
	Attributes SentenceAttributes
}

func (sent *Sentence) GetSpan() *Span {
	return &sent.Span
}
