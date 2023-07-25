package lemmatizer

import "strings"

type MorphologicalRules struct {
	NounExc map[string]string
	VerbExc map[string]string
	AdjExc  map[string]string
	AdvExc  map[string]string

	NounBase map[string]bool
	VerbBase map[string]bool
	AdjBase  map[string]bool
	AdvBase  map[string]bool
	OrdBase  map[string]bool
	CrdBase  map[string]bool

	NounRule [][]string
	VerbRule [][]string
	AdjRule  [][]string
	AbbrRule map[string]string
}

func (rules *MorphologicalRules) getNumber(form string, pos string) (string, bool) {
	if pos == "CD" {
		isOk := rules.CrdBase[form]
		if isOk {
			return "#crd#", true
		}

		isOk = rules.OrdBase[form]
		if isOk || AnyOf(form, "0st", "0nd", "0rd", "0th") {
			return "#ord#", true
		}
	}

	return "", false
}

func (rules *MorphologicalRules) getException(form string, pos string) (string, bool) {
	if IsNoun(pos) {
		exc, hasExc := rules.NounExc[form]
		return exc, hasExc
	}

	if IsVerb(pos) {
		exc, hasExc := rules.VerbExc[form]
		return exc, hasExc
	}

	if IsAdjective(pos) {
		exc, hasExc := rules.AdjExc[form]
		return exc, hasExc
	}

	if IsAdverb(pos) {
		exc, hasExc := rules.AdvExc[form]
		return exc, hasExc
	}

	return "", false
}

func (rules *MorphologicalRules) getBase(form string, pos string) (string, bool) {
	if IsNoun(pos) {
		return getBaseAux(form, rules.NounBase, rules.NounRule)
	}

	if IsVerb(pos) {
		return getBaseAux(form, rules.VerbBase, rules.VerbRule)
	}

	if IsAdjective(pos) {
		return getBaseAux(form, rules.AdjBase, rules.AdjRule)
	}

	return "", false
}

func getBaseAux(form string, set map[string]bool, rules [][]string) (string, bool) {
	for _, rule := range rules {
		if strings.HasSuffix(form, rule[0]) {
			offset := len(form) - len(rule[0])
			base := form[0:offset] + rule[1]

			isOk := set[base]
			if isOk {
				return base, true
			}
		}
	}

	return "", false
}

func (rules *MorphologicalRules) getAbbreviation(form string, pos string) (string, bool) {
	key := form + "_" + pos
	r, isOk := rules.AbbrRule[key]
	return r, isOk
}
