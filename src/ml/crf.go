package ml

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"sort"
)

type AStarNode struct {
	Parent   *AStarNode
	Viterbi  *ViterbiNode
	Cost     float64
	Priority float64
}

type ViterbiNode struct {
	InputPosition int
	OutputIndex   int
	StateID       int
	Delta         float64
	OutputState   int
}

type TransitionData struct {
	Weights       []float64 `json:"weights"`
	DefaultWeight float64   `json:"default_weight"`
}

type CRF struct {
	Features       map[string]int      `json:"features"`
	States         []string            `json:"states"`
	InitialWeights []float64           `json:"initial_weights"`
	FinalWeights   []float64           `json:"final_weights"`
	Transitions    [][]*TransitionData `json:"transitions"`
}

func (crf *CRF) DotProduct(transition *TransitionData, featureIdxVector []int) float64 {
	sort.Ints(featureIdxVector)
	transitionWeights := transition.Weights
	ret := 0.0
	for _, fIdx := range featureIdxVector {
		ret += transitionWeights[fIdx]
	}
	return ret
}

func (crf *CRF) ToFeatureIdxVector(features []Feature) []int {

	set := make(map[int]bool)
	for _, feat := range features {
		fIdx, isOk := crf.Features[feat.String()]
		if isOk {
			set[fIdx] = true
		}
	}

	result := make([]int, len(set))
	i := 0
	for k := range set {
		result[i] = k
		i++
	}
	return result
}

func (crf *CRF) AStarSearch(features [][]Feature, viterbiNodes [][]*ViterbiNode) *AStarNode {

	if len(viterbiNodes) > 0 {
		var q []*AStarNode

		finalLen := len(viterbiNodes[len(viterbiNodes)-1])
		for i := 0; i < finalLen; i++ {
			viterbiNode := viterbiNodes[len(viterbiNodes)-1][i]
			q = append(q, &AStarNode{
				Viterbi:  viterbiNode,
				Priority: -viterbiNode.Delta,
			})
		}

		alreadyProcessed := make(map[*AStarNode]bool)

		for len(q) > 0 {

			var minCostNode *AStarNode
			for _, node := range q {

				if isOk := alreadyProcessed[node]; isOk {
					continue
				}

				if minCostNode == nil {
					minCostNode = node
				} else if minCostNode.Priority > node.Priority {
					minCostNode = node
				}
			}

			if minCostNode == nil || minCostNode.Viterbi.InputPosition == 0 {
				return minCostNode
			}

			alreadyProcessed[minCostNode] = true

			vNodeLevelId := minCostNode.Viterbi.InputPosition - 1
			fIdxVector := crf.ToFeatureIdxVector(features[vNodeLevelId])

			for stateId := 0; stateId < len(crf.States); stateId++ {

				if stateId >= len(viterbiNodes[vNodeLevelId]) {
					continue
				}
				transition := crf.Transitions[stateId][minCostNode.Viterbi.StateID]
				dp := crf.DotProduct(transition, fIdxVector)
				transCost := dp + transition.DefaultWeight

				aStarNode := &AStarNode{
					Viterbi: viterbiNodes[vNodeLevelId][stateId],
					Cost:    -transCost + minCostNode.Cost,
					Parent:  minCostNode,
				}

				//cost := minCostNode.Cost + transCost
				aStarNode.Priority = -aStarNode.Viterbi.Delta + aStarNode.Cost
				q = append(q, aStarNode)

			}

		}
	}

	return nil
}

func (crf *CRF) DecodeViterbi(features [][]Feature) [][]*ViterbiNode {
	impossibleWeight := math.Inf(-1)

	nodesCnt := len(features) + 1
	nodes := make([][]*ViterbiNode, nodesCnt)

	// init weights
	nodes[0] = make([]*ViterbiNode, 0, len(crf.States))
	for i := 0; i < len(crf.States); i++ {
		if crf.InitialWeights[i] > impossibleWeight {
			vNode := ViterbiNode{
				Delta:   crf.InitialWeights[i],
				StateID: i,
			}
			nodes[0] = append(nodes[0], &vNode)
		}
	}

	for obsIdx, feats := range features {
		fIdxVector := crf.ToFeatureIdxVector(feats)
		for stateId := 0; stateId < len(crf.States); stateId++ {
			if len(nodes[obsIdx]) <= stateId || nodes[obsIdx][stateId].Delta == impossibleWeight {
				continue
			}

			transitions := crf.Transitions[stateId]

			for transId, transition := range transitions {
				dp := crf.DotProduct(transition, fIdxVector)
				transWeight := dp + transition.DefaultWeight
				weight := nodes[obsIdx][stateId].Delta + transWeight
				if obsIdx == len(features)-1 {
					weight += crf.FinalWeights[transId]
				}

				var vNode *ViterbiNode
				if nodes[obsIdx+1] != nil && transId < len(nodes[obsIdx+1]) {
					vNode = nodes[obsIdx+1][transId]
				} else {
					vNode = &ViterbiNode{
						StateID:       transId,
						InputPosition: obsIdx + 1,
						OutputIndex:   transId,
						Delta:         impossibleWeight,
					}

					nodes[obsIdx+1] = append(nodes[obsIdx+1], vNode)
				}

				if weight > vNode.Delta {
					vNode.Delta = weight
				}
			}
		}
	}

	return nodes
}

func (crf *CRF) Predict(features [][]Feature) []string {
	viterbiNodes := crf.DecodeViterbi(features)
	result := make([]string, len(features))

	node := crf.AStarSearch(features, viterbiNodes)
	if node != nil {
		node = node.Parent
	}
	i := 0
	for node != nil {
		result[i] = crf.States[node.Viterbi.OutputIndex]
		i++
		node = node.Parent
	}
	return result
}

func LoadCRFFromFile(modelPath string) (*CRF, error) {
	buf, err := ioutil.ReadFile(modelPath)
	if err != nil {
		return nil, err
	}

	var m CRF
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return nil, err
	}

	// fill absent initial weights to Infinity values
	if len(m.InitialWeights) < len(m.States) {
		for i := 0; i < len(m.States)-len(m.InitialWeights); i++ {
			m.InitialWeights = append(m.InitialWeights, math.Inf(-1))
		}
	}

	return &m, nil
}
