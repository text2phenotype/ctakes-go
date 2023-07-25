package lookup

import (
	"text2phenotype.com/fdl/utils"
	"encoding/json"
)

type RareWordTerm struct {
	Tokens        []*string
	TextLength    uint32
	CUI           *string
	RareWordIndex byte
}

func (term *RareWordTerm) GetHashCode() uint64 {
	var toHash [][]byte
	for _, t := range term.Tokens {
		toHash = append(toHash, []byte(*t))
	}

	toHash = append(toHash, []byte(*term.CUI))

	return utils.HashBytes(toHash...)
}

func (term *RareWordTerm) GetRareWord() *string {
	return term.Tokens[term.RareWordIndex]
}

func (term *RareWordTerm) GetTokenCount() int {
	return len(term.Tokens)
}

type MapListIterator func() (*RareWordTerm, bool)

func CreateMapListIterator(m map[*string][]*RareWordTerm, keys []*string) MapListIterator {
	currentKey := 0
	cursor := 0

	var actualKeys []*string
	for _, k := range keys {
		if _, hasKey := m[k]; hasKey {
			actualKeys = append(actualKeys, k)
		}
	}

	l := len(actualKeys)

	return func() (*RareWordTerm, bool) {

		if l == 0 {
			return nil, false
		}

		key := actualKeys[currentKey]

		if cursor >= len(m[key]) {
			cursor = 0
			currentKey = currentKey + 1

			if currentKey >= l {
				return nil, false
			}
			key = actualKeys[currentKey]
		}

		value := *m[key][cursor]
		cursor = cursor + 1
		return &value, true
	}
}

type RareWordTermMap map[*string][]*RareWordTerm

func (rw RareWordTermMap) MarshalJSON() ([]byte, error) {
	out := make(map[string][]*RareWordTerm)
	for k, v := range rw {
		out[*k] = v
	}
	res, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rw RareWordTermMap) UnmarshalJSON(data []byte) error {
	ss := make(map[string][]*RareWordTerm)
	err := json.Unmarshal(data, &ss)
	if err != nil {
		return err
	}

	localStore := make(map[string]*string)
	getPtr := func(s string) *string {
		ptr, isOk := localStore[s]
		if !isOk {
			ptr = utils.GlobalStringStore().GetPointer(s)
			localStore[s] = ptr
		}

		return ptr
	}

	for k, v := range ss {
		k_ptr := getPtr(k)
		for _, rWord := range v {
			rWord.CUI = getPtr(*rWord.CUI)
			for ti, token := range rWord.Tokens {
				rWord.Tokens[ti] = getPtr(*token)
			}
		}
		rw[k_ptr] = v
	}
	return nil
}
