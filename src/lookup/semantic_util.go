package lookup

import (
	"text2phenotype.com/fdl/types"
	"strings"
)

var (
	DRUG = map[string]bool{
		"T053": true, "T109": true, "T110": true, "T114": true, "T115": true, "T116": true, "T118": true, "T119": true,
		"T121": true, "T122": true, "T123": true, "T124": true, "T125": true, "T126": true, "T127": true, "T129": true,
		"T130": true, "T131": true, "T195": true, "T196": true, "T197": true, "T200": true, "T203": true,
	}

	DISO = map[string]bool{
		"T019": true, "T020": true, "T037": true, "T047": true, "T048": true, "T049": true, "T050": true, "T190": true,
		"T191": true,
	}

	FIND = map[string]bool{
		"T033": true, "T040": true, "T041": true, "T042": true, "T043": true, "T044": true, "T045": true, "T046": true,
		"T056": true, "T057": true, "T184": true,
	}

	PROC = map[string]bool{
		"T060": true, "T061": true,
	}

	ACTIVITY = map[string]bool{
		"T058": true,
	}

	LAB = map[string]bool{
		"T034": true, "T059": true, "T201": true,
	}

	ANAT = map[string]bool{
		"T021": true, "T022": true, "T023": true, "T024": true, "T025": true, "T026": true, "T029": true, "T030": true,
	}
)

func GutTUISemanticGroupID(tui string) types.Semantic {
	tui = strings.ToUpper(tui)
	if _, isDrug := DRUG[tui]; isDrug {
		return types.SemanticDrug
	}
	if _, isDrug := DISO[tui]; isDrug {
		return types.SemanticDisorder
	}
	if _, isDrug := FIND[tui]; isDrug {
		return types.SemanticFinding
	}
	if _, isDrug := PROC[tui]; isDrug {
		return types.SemanticProcedure
	}
	if _, isDrug := ACTIVITY[tui]; isDrug {
		return types.SemanticActivity
	}
	if _, isDrug := LAB[tui]; isDrug {
		return types.SemanticLab
	}
	if _, isDrug := ANAT[tui]; isDrug {
		return types.SemanticAnatomicalSite
	}

	return types.SemanticUnknown
}

func GetAspect(semantic types.Semantic) string {
	return semantic.Name()
}
