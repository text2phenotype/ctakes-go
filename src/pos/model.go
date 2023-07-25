package pos

import (
	"encoding/json"
	"io/ioutil"
	"math"
)

type Context struct {
	Outcomes   []int
	Parameters []float64
}

type EvalParameters struct {
	Params        []Context `json:"params"`
	NumOfOutcomes int       `json:"numOfOutcomes"`
}

type Model struct {
	Probs      []float64      `json:"probs"`
	Outcomes   []string       `json:"outcomes"`
	PMap       map[string]int `json:"pmap"`
	EvalParams EvalParameters `json:"evalParams"`
}

func (m Model) Eval(context []string) []float64 {
	scontexts := make([]int, len(context))
	for i := 0; i < len(context); i++ {
		ci, isOk := m.PMap[context[i]]
		if !isOk {
			ci = -1
		}

		scontexts[i] = ci
	}

	outsums := make([]float64, len(m.Probs))
	copy(outsums, m.Probs)

	params := m.EvalParams.Params
	numFeats := make([]int, m.EvalParams.NumOfOutcomes)
	var activeOutcomes []int
	var activeParameters []float64

	for _, scontext := range scontexts {
		if scontext < 0 {
			continue
		}

		predParam := params[scontext]
		activeOutcomes = predParam.Outcomes
		activeParameters = predParam.Parameters

		for ai, oid := range activeOutcomes {
			numFeats[oid]++
			outsums[oid] += activeParameters[ai]
		}
	}

	normal := 0.0
	for oid := 0; oid < m.EvalParams.NumOfOutcomes; oid++ {
		outsums[oid] = math.Exp(outsums[oid])
		normal += outsums[oid]
	}

	for oid := 0; oid < m.EvalParams.NumOfOutcomes; oid++ {
		outsums[oid] /= normal
	}

	return outsums
}

func LoadModelFromFile(modelFilePath string) (Model, error) {
	var m Model
	buf, err := ioutil.ReadFile(modelFilePath)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(buf, &m)
	return m, err
}
