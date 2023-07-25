package pipeline

import (
	"text2phenotype.com/fdl/negation"
	"text2phenotype.com/fdl/smoking"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"encoding/json"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"
)

type SmokingStatusParams struct {
	RuleBasedParams            smoking.RuleBasedClassifierParams `json:"rule_based_params"`
	PcsParams                  smoking.PcsClassifierParams       `json:"pcs_params"`
	SmokerPhrases              []string                          `json:"smoker_phrases"`
	NonSmokerPhrases           []string                          `json:"non_smoker_phrases"`
	NegationContradictionWords map[string]bool                   `json:"negation_contradiction_words"`
	Boundaries                 map[string]bool                   `json:"boundaries"`
}

func NewSmokingStatusAttributeAnnotator(params SmokingStatusParams) (func(sentChannel <-chan types.Sentence) <-chan types.Sentence, error) {

	//fdlLogger := logger.NewLogger("Smoking status attribute annotator")

	rbClassifier := smoking.NewRuleBasedClassifier(params.RuleBasedParams)
	pcsClassifier, err := smoking.NewPcsClassifier(params.PcsParams)
	if err != nil {
		return nil, err
	}

	analyzer := negation.NewPolarityAnalyzer(7, 7, params.Boundaries)

	return func(sentChannel <-chan types.Sentence) <-chan types.Sentence {

		out := make(chan types.Sentence)

		go func() {

			defer close(out)
			var wg sync.WaitGroup
			for sent := range sentChannel {
				wg.Add(1)
				go func(sent types.Sentence) {
					defer wg.Done()
					rbResult := rbClassifier(sent)
					finalClassification := smoking.ClassUnknown
					if rbResult == smoking.ClassKnown {
						pcsResult := pcsClassifier(sent)

						tokens := make([]string, len(sent.Tokens))
						for i, token := range sent.Tokens {
							tokens[i] = *token.Text
						}

						smokerAnnotation := searchSmokingAnnotation(sent, params.SmokerPhrases)
						polarities, _ := analyzer(smokerAnnotation, []types.Scope{types.ScopeLeft, types.ScopeRight})
						negCnt := 0
						for _, polarity := range polarities {
							if polarity == types.PolarityNegative {
								negCnt++
							}
						}
						nonSmokerAnnotation := searchSmokingAnnotation(sent, params.NonSmokerPhrases)

						negConCnt := getNegConCount(sent.Tokens, params.NegationContradictionWords)

						if (negCnt > 0 && negConCnt == 0) || len(nonSmokerAnnotation) > 0 {
							finalClassification = smoking.ClassNonSmoker
						} else {
							finalClassification = pcsResult
						}
					}

					sent.Attributes.SmokingStatus = finalClassification
					out <- sent
				}(sent)

			}

			wg.Wait()
		}()

		return out

	}, nil
}

func searchSmokingAnnotation(sent types.Sentence, dict []string) []types.Annotation {
	result := make([]types.Annotation, 0)
	sentText := strings.ToLower(*sent.Text)
	for _, record := range dict {
		recLen := utf8.RuneCountInString(record)
		window := sentText
		offset := 0

		i := strings.Index(window, record)
		for i != -1 {
			offset += int(sent.Begin) + i
			result = append(result, types.Annotation{
				Sentence: &sent,
				Span: types.Span{
					Begin: int32(offset),
					End:   int32(offset + recLen),
					Text:  &record,
				},
			})
			window = window[i+len(record):]
			i = strings.Index(window, record)
		}
	}

	return result
}

func getNegConCount(tokens []*types.Token, negationWords map[string]bool) int {
	replacePtr := regexp.MustCompile(`[\W]`)
	conCnt := 0

	for _, token := range tokens {
		tok := strings.TrimSpace(replacePtr.ReplaceAllString(*token.Text, " "))
		toks := strings.Split(tok, " ")
		for i := 0; i < len(toks); i++ {
			if _, isOk := negationWords[toks[i]]; isOk {
				conCnt++
			}
		}
	}
	return conCnt
}

func LoadSmokingStatusParameters(smokingResPath string) (SmokingStatusParams, error) {
	var res SmokingStatusParams

	// load PCS classifier params
	// load PCS model
	modelPath := path.Join(smokingResPath, "PCS", "pcs_libsvm.model.json")
	buf, err := ioutil.ReadFile(modelPath)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(buf, &res.PcsParams.Model)
	if err != nil {
		return res, err
	}

	// load stop words
	stopWordPath := path.Join(smokingResPath, "PCS", "stopwords_PCS.txt")
	res.PcsParams.StopWords, err = utils.ReadSet(stopWordPath)
	if err != nil {
		return res, err
	}

	// load go words
	goWordPath := path.Join(smokingResPath, "PCS", "keywords_PCS.txt")
	res.PcsParams.GoWords, err = utils.ReadList(goWordPath)
	if err != nil {
		return res, err
	}

	// load rule base classifier params
	// load unk words
	unkWordPath := path.Join(smokingResPath, "KU", "unknown_words.txt")
	res.RuleBasedParams.UnknownWords, err = utils.ReadList(unkWordPath)
	if err != nil {
		return res, err
	}

	// load go words
	keyWordPath := path.Join(smokingResPath, "KU", "keywords.txt")
	res.RuleBasedParams.SmokingWords, err = utils.ReadSet(keyWordPath)
	if err != nil {
		return res, err
	}

	// load smoking dicts
	smokerDictPath := path.Join(smokingResPath, "smoker.dictionary")
	res.SmokerPhrases, err = utils.ReadList(smokerDictPath)
	if err != nil {
		return res, err
	}

	nonSmokerDictPath := path.Join(smokingResPath, "nonsmoker.dictionary")
	res.NonSmokerPhrases, err = utils.ReadList(nonSmokerDictPath)
	if err != nil {
		return res, err
	}

	// load neg words
	negWordsPath := path.Join(smokingResPath, "context", "negationContradictionWords.txt")
	res.NegationContradictionWords, err = utils.ReadSet(negWordsPath)
	if err != nil {
		return res, err
	}

	// load boundaries
	boundariesPath := path.Join(smokingResPath, "context", "boundaryData.txt")
	res.Boundaries, err = utils.ReadSet(boundariesPath)
	if err != nil {
		return res, err
	}

	return res, err
}
