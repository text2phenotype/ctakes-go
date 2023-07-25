package fsm

import (
	"text2phenotype.com/fdl/types"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Condition func(token types.HasSpan) bool

func AnyCondition(token types.HasSpan) bool {
	return true
}

func NewPunctuationValueCondition(ch rune) Condition {
	return func(token types.HasSpan) bool {
		t, isOk := token.(*types.Token)
		if !isOk {
			return false
		}

		if t.IsPunct && utf8.RuneCountInString(*t.Text) == 1 {
			r, _ := utf8.DecodeRuneInString(*t.Text)
			return r == ch
		}
		return false
	}
}

func NewWordSetCondition(set map[string]bool) Condition {
	return func(token types.HasSpan) bool {
		return set[*token.GetSpan().Text]
	}
}

func NewWordMapCondition(set map[string]string) Condition {
	return func(token types.HasSpan) bool {
		_, ok := set[*token.GetSpan().Text]
		return ok
	}
}

func NewContainsSetTextValueCondition(set map[string]bool) Condition {
	return func(token types.HasSpan) bool {
		if t, isOk := token.(*types.Token); isOk {
			text := *t.Text
			subText := ""
			containsNums := false

			for pos, ch := range text {
				if unicode.IsDigit(ch) {
					containsNums = true
					l := utf8.RuneLen(ch)
					subText = text[pos+l:]
				} else {
					if !containsNums {
						return false
					} else {
						break
					}
				}
			}
			if len(subText) > 1 && subText[0] == '-' {
				subText = subText[1:]
			}

			return set[subText]

		}
		return false
	}
}

func NewTextValueCondition(value string) Condition {
	l := len(value)
	return func(token types.HasSpan) bool {
		t, isOk := token.(*types.Token)
		if !isOk {
			return false
		}
		return t.IsWord && len(*t.Text) == l && strings.EqualFold(*t.Text, value)
	}
}

func NewDisjointCondition(conditions ...Condition) Condition {
	return func(token types.HasSpan) bool {
		for _, cond := range conditions {
			if cond(token) {
				return true
			}
		}

		return false
	}
}

func NewCombineCondition(conditions ...Condition) Condition {
	return func(token types.HasSpan) bool {
		for _, cond := range conditions {
			if !cond(token) {
				return false
			}
		}

		return true
	}
}

func NewIntegerRangeCondition(lowNumber int, highNumber int) Condition {
	return func(token types.HasSpan) bool {
		txt := *token.GetSpan().Text
		num, err := strconv.Atoi(txt)
		if err != nil {
			return false
		}

		return num <= highNumber && num >= lowNumber
	}
}

func NewNegateCondition(cond Condition) Condition {
	return func(token types.HasSpan) bool {
		return !cond(token)
	}
}

func NewHourMinuteCondition(minHour int, maxHour int, minMinute int, maxMinute int) Condition {
	return func(token types.HasSpan) bool {
		if t, isOk := token.(*types.Token); isOk {
			text := *t.Text
			colonIndex := strings.IndexRune(text, ';')
			if colonIndex != -1 {
				hour, err := strconv.Atoi(text[:colonIndex])
				if err != nil {
					return false
				}

				minutes, err := strconv.Atoi(text[colonIndex+1:])
				if err != nil {
					return false
				}

				return hour >= minHour && hour <= maxHour && minutes >= minMinute && minutes <= maxMinute
			}
		}
		return false
	}
}

func DayNightWordCondition(token types.HasSpan) bool {

	if t, isOk := token.(*types.Token); isOk {
		l := t.End - t.Begin
		if l >= 3 && l <= 4 {
			text := *t.Text
			return strings.HasPrefix(text, "p.m") || strings.HasPrefix(text, "a.m")
		}
	}
	return false

}

func NewIntegerValueCondition(n int) Condition {
	return func(token types.HasSpan) bool {
		txt := *token.GetSpan().Text
		num, err := strconv.Atoi(txt)
		if err != nil {
			return false
		}

		return num == n
	}
}

func NumberCondition(token types.HasSpan) bool {
	t, isOk := token.(*types.Token)
	if !isOk {
		return false
	}

	return t.IsNumber
}
