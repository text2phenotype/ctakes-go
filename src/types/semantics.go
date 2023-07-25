package types

type Semantic byte

func (s Semantic) Name() string {
	switch s {
	case SemanticDrug:
		return "drug"
	case SemanticDisorder:
		return "prob"
	case SemanticFinding:
		return "symp"
	case SemanticProcedure:
		return "proc"
	case SemanticAnatomicalSite:
		return "anat"
	case SemanticDevice:
		return "device"
	case SemanticLab:
		return "lab"
	case SemanticPhenomena:
		return "pheno"
	case SemanticActivity:
		return "activity"
	}
	return "unknown"
}

const (
	SemanticUnknown        Semantic = 0
	SemanticDrug           Semantic = 1
	SemanticDisorder       Semantic = 2
	SemanticFinding        Semantic = 3
	SemanticProcedure      Semantic = 5
	SemanticAnatomicalSite Semantic = 6
	SemanticDevice         Semantic = 8
	SemanticLab            Semantic = 9
	SemanticPhenomena      Semantic = 10
	SemanticActivity       Semantic = 11
)
