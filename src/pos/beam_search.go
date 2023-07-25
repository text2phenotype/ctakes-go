package pos

import (
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"container/heap"
	"sort"
)

const minSequenceScore = -100000

func NewBeamSearch(model Model, size int) func(sequence []*types.Token, contentGen ContextGenerator, sequenceValidator SequenceValidator) (Sequence, bool) {

	return func(sequence []*types.Token, contentGen ContextGenerator, sequenceValidator SequenceValidator) (Sequence, bool) {
		prev := make(utils.PriorityQueue, 0, size)
		heap.Init(&prev)
		next := make(utils.PriorityQueue, 0, size)
		heap.Init(&next)
		heap.Push(&prev, Sequence{})

		for i := 0; i < len(sequence); i++ {
			sz := len(prev)
			if size < sz {
				sz = size
			}

			for sc := 0; len(prev) > 0 && sc < sz; sc++ {
				top := heap.Pop(&prev).(Sequence)
				outcomes := top.Outcomes

				//outcomes := make([]string, len(tmpOutcomes))
				contexts := contentGen.GetContext(i, sequence, outcomes)
				scores := model.Eval(contexts)

				tempScores := make([]float64, len(scores))
				copy(tempScores, scores)
				sort.Float64s(tempScores)

				idx := len(scores) - size
				if idx < 0 {
					idx = 0
				}
				min := tempScores[idx]

				for p := 0; p < len(scores); p++ {
					if scores[p] < min {
						continue
					}

					out := model.Outcomes[p]
					if sequenceValidator.ValidSequence(i, sequence, out) {
						var ns Sequence
						ns.ExpandFrom(top, out, scores[p])
						if ns.Score > minSequenceScore {
							heap.Push(&next, ns)
						}
					}
				}

				if len(next) == 0 {
					for p := 0; p < len(scores); p += 1 {
						out := model.Outcomes[p]
						if sequenceValidator.ValidSequence(i, sequence, out) {
							var ns Sequence
							ns.ExpandFrom(top, out, scores[p])
							if ns.Score > minSequenceScore {
								heap.Push(&next, ns)
							}
						}
					}
				}

			}

			prev = utils.PriorityQueue{}
			heap.Init(&prev)
			prev, next = next, prev
		}

		var topSequence Sequence
		isOk := false

		if len(prev) > 0 {
			topSequence = heap.Pop(&prev).(Sequence)
			isOk = true
		}

		return topSequence, isOk
	}
}
