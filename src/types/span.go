package types

import (
	"text2phenotype.com/fdl/utils"
	"fmt"
)

type HasSpan interface {
	GetSpan() *Span
}

type Span struct {
	Begin int32
	End   int32
	Text  *string
}

func CheckSpansOverlap(covered *Span, covering *Span) bool {
	return covering.Begin <= covered.Begin && covering.End >= covered.End
}

func (span Span) GetHashCode() uint64 {

	key := fmt.Sprintf("%d_%d", span.Begin, span.End)
	return utils.HashString(key)
}

func (span Span) GetTextFromSentence(sent *Sentence) (string, bool) {
	if span.Begin < sent.Begin || span.End > sent.End {
		return "", false
	}

	localBegin := span.Begin - sent.Begin
	localEnd := span.End - sent.Begin

	runes := []rune(*sent.Text)
	return string(runes[localBegin:localEnd]), true
}

type Spans []HasSpan

func (spans Spans) Len() int {
	return len(spans)
}

func (spans Spans) Less(i int, j int) bool {
	spanI, spanJ := spans[i].GetSpan(), spans[j].GetSpan()

	if spanI.Begin == spanJ.Begin {
		return spanI.End < spanJ.End
	}
	return spanI.Begin < spanJ.Begin
}

func (spans Spans) Swap(i int, j int) {
	spans[i], spans[j] = spans[j], spans[i]
}

func (spans Spans) SearchFirstInSpan(begin int32, end int32) (int32, bool) {
	for i, span := range spans {
		if span.GetSpan().Begin >= begin && span.GetSpan().End <= end {
			return int32(i), true
		}
	}
	return -1, false
}

func SpanSortFunction(spanA *Span, spanB *Span) bool {
	if spanA.Begin == spanB.Begin {
		return spanA.End < spanB.End
	}
	return spanA.Begin < spanB.Begin
}
