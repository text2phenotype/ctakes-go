package pipeline

import "text2phenotype.com/fdl/types"

type RequestVocabParams struct {
	Name   string              `json:"name"`
	Params types.RequestParams `json:"params"`
}

type Request struct {
	Text string `json:"redis_key"`
	Tid  string `json:"tid"`
}
