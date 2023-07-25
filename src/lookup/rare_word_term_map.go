package lookup

import (
	"math"
	"regexp"
	"sync"
)

// hashset alternative
var badPosTerms = map[string]bool{
	"about": true, "across": true, "after": true, "against": true, "all": true, "along": true, "and": true,
	"any": true, "around": true, "at": true, "away": true, "back": true, "before": true, "behind": true,
	"below": true, "beneath": true, "beside": true, "besides": true, "between": true, "beyond": true, "both": true,
	"but": true, "by": true, "can": true, "concerning": true, "could": true, "down": true, "during": true,
	"eight": true, "except": true, "five": true, "for": true, "forward": true, "four": true, "from": true,
	"half": true, "he": true, "hers": true, "his": true, "how": true, "however": true, "i": true, "in": true,
	"inside": true, "into": true, "it": true, "its": true, "like": true, "may": true, "might": true, "mine": true,
	"must": true, "my": true, "near": true, "nine": true, "none": true, "nor": true, "of": true, "off": true,
	"on": true, "one": true, "or": true, "our": true, "ours": true, "out": true, "outside": true, "over": true,
	"past": true, "seven": true, "she": true, "should": true, "since": true, "six": true, "so": true, "some": true,
	"ten": true, "that": true, "the": true, "theirs": true, "there": true, "these": true, "this": true, "those": true,
	"three": true, "through": true, "throughout": true, "to": true, "toward": true, "twice": true, "two": true,
	"under": true, "until": true, "up": true, "upon": true, "what": true, "whatever": true, "when": true,
	"whenever": true, "where": true, "wherever": true, "which": true, "whichever": true, "who": true, "whoever": true,
	"whom": true, "whomever": true, "will": true, "with": true, "without": true, "would": true, "yet": true,
	"you": true, "yours": true, "zero": true,
}

var hasLetterPtr *regexp.Regexp
var hasLetterOnce sync.Once

func isRarableToken(token *string) bool {
	if len(*token) <= 1 {
		return false
	}

	hasLetterOnce.Do(func() {
		pttr, _ := regexp.Compile("[a-zA-Z]+")
		hasLetterPtr = pttr
	})

	// check if at least one letter exists
	loc := hasLetterPtr.FindStringIndex(*token)

	if len(loc) == 0 {
		return false
	}

	_, ok := badPosTerms[*token]
	return !ok
}

func createRareWordMap(rareWords []*RareWordTerm) RareWordTermMap {
	countMap := createTokenCountMap(rareWords)

	termMap := make(RareWordTermMap)
	for _, term := range rareWords {
		fillRareWord(term, &countMap)
		rareWordHash := term.GetRareWord()

		rareWordTerms, ok := termMap[rareWordHash]
		if !ok {
			rareWordTerms = make([]*RareWordTerm, 1)
			rareWordTerms[0] = term
		} else {
			rareWordTerms = append(rareWordTerms, term)
		}

		termMap[rareWordHash] = rareWordTerms
	}
	return termMap
}

func createTokenCountMap(rare_words []*RareWordTerm) map[*string]int {
	count_map := make(map[*string]int)

	for _, term := range rare_words {
		for _, token := range term.Tokens {
			if isRarableToken(token) {
				count, ok := count_map[token]
				if !ok {
					count = 0
				}
				count_map[token] = count + 1
			}
		}
	}

	return count_map
}

func fillRareWord(term *RareWordTerm, countMap *map[*string]int) {
	if len(term.Tokens) <= 1 {
		return
	}

	var rareIndex byte = 0
	var minCount = math.MaxInt32
	for i, word := range term.Tokens {
		cnt, ok := (*countMap)[word]
		if !ok {
			continue
		}

		if cnt < minCount {
			minCount = cnt
			rareIndex = byte(i)
		}
	}

	term.RareWordIndex = rareIndex
}
