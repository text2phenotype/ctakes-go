package types

type BaseResponse struct {
	DocId  string `json:"docId"`
	Dob    string `json:"dob"`
	Gender string `json:"gender"`
	Age    string `json:"age"`
}

type ContentSection struct {
	Id            int                    `json:"id"`
	Sentence      []int32                `json:"sentence"`
	SectionOffset []int32                `json:"sectionOffset"`
	Text          []interface{}          `json:"text"`
	SectionOid    string                 `json:"sectionOid"`
	Attributes    map[string]interface{} `json:"attributes"`
	Aspect        string                 `json:"aspect"`
	Name          string                 `json:"name"`
	UmlsConcepts  []UmlsConcept          `json:"umlsConcepts"`
}

type UmlsConcept struct {
	Tui           []string     `json:"tui"`
	Cui           string       `json:"cui"`
	PreferredText string       `json:"preferredText"`
	SabConcepts   []SabConcept `json:"sabConcepts"`
}

type SabConcept struct {
	CodingScheme  string         `json:"codingScheme"`
	VocabConcepts []VocabConcept `json:"vocabConcepts"`
}

type VocabConcept struct {
	Tty  []string `json:"tty"`
	Code string   `json:"code"`
}

type DefaultClinicalResponse struct {
	BaseResponse
	Content []ContentSection `json:"content"`
}

type SmokingStatusSection struct {
	Status string
	Text   []interface{}
}

type SmokingStatusResponse struct {
	BaseResponse
	SmokingStatus string                 `json:"smokingStatus"`
	Sentences     []SmokingStatusSection `json:"sentences"`
}
