package types

import (
	"strings"
)

type ConceptCodes map[string]CodeParams // code name -> []params

type CodeParams map[string][]string // param name -> []values

type Concept struct {
	CUI       *string
	PrefText  string
	Semantics []Semantic
	TUI       []string
	Codes     map[string]ConceptCodes // sab -> codes
}

func CreateConcept(columns []string, cui *string, schemeMap map[string]int) *Concept {
	return &Concept{
		CUI:       cui,
		PrefText:  columns[schemeMap[PREF]],
		Semantics: []Semantic{SemanticUnknown},
		Codes:     make(map[string]ConceptCodes),
	}
}

func (concept *Concept) Update(tui string, columns []string, schemeMap map[string]int) {

	tuiIsExist := false
	for _, conceptTui := range concept.TUI {
		if strings.EqualFold(conceptTui, tui) {
			tuiIsExist = true
			break
		}
	}

	if !tuiIsExist {
		concept.TUI = append(concept.TUI, tui)
	}

	sabIdx := schemeMap[SAB]
	sab := columns[sabIdx]
	if len(sab) > 0 {
		codeIdx := schemeMap[CODE]
		code := columns[codeIdx]

		codes, ok := concept.Codes[sab]
		if !ok {
			codes = make(ConceptCodes)
		}

		params, ok := codes[code]
		if !ok {
			params = make(CodeParams)
		}

		for paramName, paramIdx := range schemeMap {

			// skip required columns
			if paramName == CUI ||
				paramName == TUI ||
				paramName == CODE ||
				paramName == SAB ||
				paramName == PREF {
				continue
			}
			newParamValue := columns[paramIdx]
			paramNamePtr := paramName

			param, hasParam := params[paramNamePtr]
			if !hasParam {
				param = []string{}
			}

			hasValue := false
			for _, paramValue := range param {
				if paramValue == newParamValue {
					hasValue = true
					break
				}
			}

			if !hasValue {
				param = append(param, newParamValue)
			}

			params[paramNamePtr] = param

		}

		codes[code] = params
		concept.Codes[sab] = codes
	}
}
