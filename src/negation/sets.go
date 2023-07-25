package negation

func getModalVerbs() map[string]bool {
	return map[string]bool{
		"can":    true,
		"ca":     true,
		"will":   true,
		"must":   true,
		"could":  true,
		"would":  true,
		"should": true,
		"shall":  true,
		"did":    true,
	}
}

func getNegParticles() map[string]bool {
	return map[string]bool{
		"not": true,
		"n't": true,
		"'t":  true,
	}
}

func getNegColPart() map[string]bool {
	return map[string]bool{
		"out": true,
	}
}

func getNegColloc() map[string]bool {
	return map[string]bool{
		"rule":     true,
		"rules":    true,
		"ruled":    true,
		"ruling":   true,
		"rule-out": true,
	}
}

func getRegVerbs() map[string]bool {
	return map[string]bool{
		"reveal":        true,
		"reveals":       true,
		"revealed":      true,
		"revealing":     true,
		"have":          true,
		"had":           true,
		"has":           true,
		"feel":          true,
		"feels":         true,
		"felt":          true,
		"feeling":       true,
		"complain":      true,
		"complains":     true,
		"complained":    true,
		"complaining":   true,
		"demonstrate":   true,
		"demonstrates":  true,
		"demonstrated":  true,
		"demonstrating": true,
		"appear":        true,
		"appears":       true,
		"appeared":      true,
		"appearing":     true,
		"caused":        true,
		"cause":         true,
		"causing":       true,
		"causes":        true,
		"find":          true,
		"finds":         true,
		"found":         true,
		"discover":      true,
		"discovered":    true,
		"discovers":     true,
	}
}

func getNegVerbs() map[string]bool {
	return map[string]bool{
		"deny":      true,
		"denies":    true,
		"denied":    true,
		"denying":   true,
		"fail":      true,
		"fails":     true,
		"failed":    true,
		"failing":   true,
		"decline":   true,
		"declines":  true,
		"declined":  true,
		"declining": true,
		"exclude":   true,
		"excludes":  true,
		"excluding": true,
		"excluded":  true,
	}
}

func getNegPrepositions() map[string]bool {
	return map[string]bool{
		"without": true,
		"absent":  true,
		"none":    true,
	}
}

func getNegDeterminers() map[string]bool {
	return map[string]bool{
		"no":      true,
		"any":     true,
		"neither": true,
		"nor":     true,
		"never":   true,
	}
}

func getRegNouns() map[string]bool {
	return map[string]bool{
		"evidence":    true,
		"indication":  true,
		"indications": true,
		"sign":        true,
		"signs":       true,
		"symptoms":    true,
		"symptom":     true,
		"sx":          true,
		"dx":          true,
		"diagnosis":   true,
		"history":     true,
		"hx":          true,
		"findings":    true,
	}
}

func getNegAdjectives() map[string]bool {
	return map[string]bool{
		"unremarkable": true,
		"unlikely":     true,
		"negative":     true,
		"no":           true,
		"unclear":      true,
	}
}
