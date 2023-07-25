package fsm

import (
	"text2phenotype.com/fdl/fsm"
	"text2phenotype.com/fdl/types"
	"strconv"
	"strings"
)

func FormCondition(token types.HasSpan) bool {
	_, isOK := token.(*FormToken)
	return isOK
}

func RangeCondition(token types.HasSpan) bool {
	_, isOK := token.(*RangeToken)
	return isOK
}

func RangeStrengthCondition(token types.HasSpan) bool {
	_, isOK := token.(*RangeStrengthToken)
	return isOK
}

func FractionStrengthCondition(token types.HasSpan) bool {
	_, isOK := token.(*FractionStrengthToken)
	return isOK
}

func RouteCondition(token types.HasSpan) bool {
	_, isOK := token.(*RouteToken)
	return isOK
}

func StrengthCondition(token types.HasSpan) bool {
	_, isOK := token.(*StrengthToken)
	return isOK
}

func FrequencyUnitCondition(token types.HasSpan) bool {
	_, isOK := token.(*FrequencyUnitToken)
	return isOK
}

func StrengthUnitCombinedCondition(token types.HasSpan) bool {
	_, isOK := token.(*StrengthUnitCombinedToken)
	return isOK
}

func StrengthUnitCondition(token types.HasSpan) bool {
	_, isOK := token.(*StrengthUnitToken)
	return isOK
}

func TimeCondition(token types.HasSpan) bool {
	_, isOK := token.(*TimeToken)
	return isOK
}

//func DecimalCondition(token types.HasSpan) bool {
//	_, isOK := token.(DecimalStrengthToken)
//	return isOK
//}

func NewHourMinuteCondition(minHour int, maxHour int, minMinute int, maxMinute int) fsm.Condition {
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

func NewIntegerValueCondition(n int) fsm.Condition {
	return func(token types.HasSpan) bool {
		txt := *token.GetSpan().Text
		num, err := strconv.Atoi(txt)
		if err != nil {
			return false
		}

		return num == n
	}
}
