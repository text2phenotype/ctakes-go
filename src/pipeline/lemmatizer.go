package pipeline

import (
	"bufio"
	"text2phenotype.com/fdl/lemmatizer"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"errors"
	"os"
	"path"
	"strings"
	"sync"
)

func NewLemmatizer(resPath string) (func(in <-chan types.Sentence) <-chan types.Sentence, error) {
	abbrRulePath := path.Join(resPath, "abbr_rule.bsv")
	adjRulePath := path.Join(resPath, "adj_rule.bsv")
	nounRulePath := path.Join(resPath, "noun_rule.bsv")
	verbRulePath := path.Join(resPath, "verb_rule.bsv")

	adjExcPath := path.Join(resPath, "adj_exc.bsv")
	advExcPath := path.Join(resPath, "adv_exc.bsv")
	nounExcPath := path.Join(resPath, "noun_exc.bsv")
	verbExcPath := path.Join(resPath, "verb_exc.bsv")

	adjBasePath := path.Join(resPath, "adj_base.txt")
	advBasePath := path.Join(resPath, "adv_base.txt")
	crdBasePath := path.Join(resPath, "crd_base.txt")
	ordBasePath := path.Join(resPath, "ord_base.txt")
	nounBasePath := path.Join(resPath, "noun_base.txt")
	verbBasePath := path.Join(resPath, "verb_base.txt")

	var rules lemmatizer.MorphologicalRules
	var err error
	if rules.AbbrRule, err = utils.ReadMap(abbrRulePath); err != nil {
		return nil, err
	}
	if rules.AdjRule, err = ReadRuleList(adjRulePath); err != nil {
		return nil, err
	}
	if rules.NounRule, err = ReadRuleList(nounRulePath); err != nil {
		return nil, err
	}
	if rules.VerbRule, err = ReadRuleList(verbRulePath); err != nil {
		return nil, err
	}
	if rules.NounExc, err = utils.ReadMap(nounExcPath); err != nil {
		return nil, err
	}
	if rules.VerbExc, err = utils.ReadMap(verbExcPath); err != nil {
		return nil, err
	}
	if rules.AdjExc, err = utils.ReadMap(adjExcPath); err != nil {
		return nil, err
	}
	if rules.AdvExc, err = utils.ReadMap(advExcPath); err != nil {
		return nil, err
	}
	if rules.NounBase, err = utils.ReadSet(nounBasePath); err != nil {
		return nil, err
	}
	if rules.VerbBase, err = utils.ReadSet(verbBasePath); err != nil {
		return nil, err
	}
	if rules.AdjBase, err = utils.ReadSet(adjBasePath); err != nil {
		return nil, err
	}
	if rules.AdvBase, err = utils.ReadSet(advBasePath); err != nil {
		return nil, err
	}
	if rules.OrdBase, err = utils.ReadSet(ordBasePath); err != nil {
		return nil, err
	}
	if rules.CrdBase, err = utils.ReadSet(crdBasePath); err != nil {
		return nil, err
	}

	analyzer, err := lemmatizer.NewMorphologicalAnalyzer(&rules)
	if err != nil {
		return nil, err
	}

	return func(in <-chan types.Sentence) <-chan types.Sentence {
		out := make(chan types.Sentence)

		go func() {
			stringStore := utils.GlobalStringStore()
			defer close(out)
			var wg sync.WaitGroup
			for sent := range in {
				wg.Add(1)

				go func(sent types.Sentence) {
					defer wg.Done()
					if sent.Tokens != nil || len(sent.Tokens) > 0 {
						for _, token := range sent.Tokens {
							if !token.IsWord || token.Text == nil || token.Tag == nil {
								continue
							}

							token.Lemma = stringStore.GetPointer(analyzer(*token.Text, *token.Tag))
						}
					}

					out <- sent
				}(sent)

			}
			wg.Wait()
		}()
		return out
	}, nil
}

func ReadRuleList(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var result [][]string
	for scanner.Scan() {
		p := strings.Split(scanner.Text(), "|")
		if len(p) != 2 {
			return nil, errors.New("rule should has 2 columns")
		}
		result = append(result, p)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
