package types

type Polarity int8

func (p Polarity) Name() string {
	switch p {
	case PolarityPositive:
		return "positive"
	case PolarityNegative:
		return "negative"
	default:
		return "neutral"
	}
}

const (
	PolarityPositive Polarity = 1
	PolarityNegative Polarity = -1
	PolarityNeutral  Polarity = 0
)

type Scope int8

const (
	ScopeLeft   Scope = -1
	ScopeMiddle Scope = 0
	ScopeRight  Scope = 1
)
