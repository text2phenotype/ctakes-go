package pos

import "math"

type Sequence struct {
	Score    float64
	Outcomes []string
	Probs    []float64
}

func (seq *Sequence) ExpandFrom(src Sequence, out string, score float64) {
	seq.Outcomes = make([]string, len(src.Outcomes)+1)
	copy(seq.Outcomes, src.Outcomes)
	seq.Outcomes[len(seq.Outcomes)-1] = out

	seq.Probs = make([]float64, len(src.Probs)+1)
	copy(seq.Probs, src.Probs)
	seq.Probs[len(seq.Probs)-1] = score

	seq.Score = src.Score + math.Log(score)
}

func (seq Sequence) Less(o interface{}) bool {
	c, isOk := o.(Sequence)
	if isOk {
		return seq.Score > c.Score
	}
	return false
}
