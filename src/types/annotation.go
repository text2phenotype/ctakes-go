package types

type Annotation struct {
	Span
	Semantic   Semantic
	Concepts   []*Concept
	Sentence   *Sentence
	Attributes map[string]interface{}
}

func (ann Annotation) GetName() string {
	switch ann.Semantic {
	case SemanticDrug:
		return "MedicationMention"
	case SemanticDisorder:
		return "DiseaseDisorderMention"
	case SemanticProcedure:
		return "ProcedureMention"
	case SemanticAnatomicalSite:
		return "AnatomicalSiteMention"
	case SemanticLab:
		return "LabMention"
	case SemanticActivity:
		return "ActivityMention"
	case SemanticFinding:
		return "SignSymptomMention"
	default:
		return "EntityMention"
	}
}

func (ann *Annotation) GetSpan() *Span {
	return &ann.Span
}
