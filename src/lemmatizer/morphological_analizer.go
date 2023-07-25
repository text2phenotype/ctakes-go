package lemmatizer

import "strings"

type MorphologicalAnalyzer func(form string, pos string) string

func NewMorphologicalAnalyzer(rules *MorphologicalRules) (MorphologicalAnalyzer, error) {
	lib, err := NewMTLib()
	if err != nil {
		return nil, err
	}

	return func(form string, pos string) string {
		form = lib.normalizeBasic(form)
		form = strings.ToLower(form)
		pos = strings.ToUpper(pos)

		number, isNumber := rules.getNumber(form, pos)
		if isNumber {
			return number
		}

		// exceptions
		exception, isExeption := rules.getException(form, pos)
		if isExeption {
			return exception
		}

		// base-forms
		base, isBase := rules.getBase(form, pos)
		if isBase {
			return base
		}

		// abbreviations
		abbreviation, isAbbreviation := rules.getAbbreviation(form, pos)
		if isAbbreviation {
			return abbreviation
		}

		return form

	}, nil
}
