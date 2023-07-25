package model

import (
	"encoding/json"
	"io/ioutil"
)

type Model struct {
	Bias          float64               `json:"bias"`
	W             []float64             `json:"weights"`
	Labels        []byte                `json:"labels"`
	FeaturesLen   int                   `json:"features_len"`
	FeaturesCache map[string]*CacheNode `json:"features_cache"`
}

type FeatureValue struct {
	Index int
	Value float64
}

type Features interface {
	Values() []FeatureValue // feature Index -> feature Value
}

type CacheNode struct {
	Value int                   `json:"v"`
	Inner map[string]*CacheNode `json:"i"`
}

func Load(path string) (*Model, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Model
	err = json.Unmarshal(buf, &m)
	return &m, err
}

func (model *Model) Predict(x Features) byte {
	decValues := make([]float64, len(model.Labels))

	n := model.FeaturesLen
	if model.Bias >= 0.0 {
		n = n + 1
	}

	nrW := len(model.Labels)
	if nrW == 2 {
		nrW = 1
	}

	var decMaxIdx int
	for decMaxIdx = 0; decMaxIdx < nrW; decMaxIdx++ {
		decValues[decMaxIdx] = 0
	}

	featValues := x.Values()
	for _, featValue := range featValues {
		idx, value := featValue.Index, featValue.Value
		if idx <= n {
			for i := 0; i < nrW; i++ {
				decValues[i] += model.W[(idx-1)*nrW+i] * value
			}
		}
	}

	if len(model.Labels) == 2 {
		lblIdx := 1
		if decValues[0] > 0.0 {
			lblIdx = 0
		}
		return model.Labels[lblIdx]
	} else {
		decMaxIdx = 0
		for i := 0; i < len(model.Labels); i++ {
			if decValues[i] > decValues[decMaxIdx] {
				decMaxIdx = i
			}
		}
		return model.Labels[decMaxIdx]
	}
}

func (model *Model) GetFeatureIndex(featureNameParts []string) (int, bool) {
	node := &CacheNode{Inner: model.FeaturesCache}
	for _, featureNamePart := range featureNameParts {
		var ok bool
		node, ok = node.Inner[featureNamePart]
		if !ok {
			return 0, false
		}
	}
	if node.Value < 0 {
		return 0, false
	}
	return node.Value, true
}
