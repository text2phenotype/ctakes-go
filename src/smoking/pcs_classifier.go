package smoking

import (
	"text2phenotype.com/fdl/ml/svm"
	"text2phenotype.com/fdl/types"
	"regexp"
	"strings"
)

type PcsClassifierParams struct {
	StopWords map[string]bool `json:"stop_words"`
	GoWords   []string        `json:"go_words"`
	Model     svm.Model       `json:"model"`
}

func NewPcsClassifier(params PcsClassifierParams) (func(sent types.Sentence) string, error) {

	dashSplit, err := regexp.Compile("-{2,}")
	if err != nil {
		return nil, err
	}

	datePtr := []*regexp.Regexp{
		regexp.MustCompile(`19\d\d`),
		regexp.MustCompile(`19\d\ds`),
		regexp.MustCompile(`20\d\d`),
		regexp.MustCompile(`20\d\ds`),
		regexp.MustCompile(`[1-9]0s`),
		regexp.MustCompile(`\d{1,2}[/-]\d{1,2}`),
		regexp.MustCompile(`\d{1,2}[/-]\d{4}`),
		regexp.MustCompile(`\d{1,2}[/-]\d{1,2}[/-]\d{2}`),
		regexp.MustCompile(`\d{1,2}[/-]\d{1,2}[/-]\d{4}`),
	}

	sentSymbolsReplacePtr, err := regexp.Compile(`[.?!:;()',"{}<>#+]`)
	if err != nil {
		return nil, err
	}

	return func(sent types.Sentence) string {

		var unigrams []string
		for _, token := range sent.Tokens {
			if !token.IsWord {
				continue
			}
			tokenText := strings.TrimSpace(strings.ToLower(*token.Text))
			tokenText = dashSplit.ReplaceAllString(tokenText, " ")
			toks := strings.Split(tokenText, " ")
			for i := 0; i < len(toks); i++ {
				if isOk := params.StopWords[toks[i]]; !isOk {
					unigrams = append(unigrams, toks[i])
				}
			}
		}

		bigrams := make([]string, 0, len(unigrams)-1)
		for i := 0; i < len(unigrams)-1; i++ {
			var sb strings.Builder
			sb.WriteString(unigrams[i])
			sb.WriteRune('_')
			sb.WriteString(unigrams[i+1])
			bigrams = append(bigrams, sb.String())
		}

		var features []float64
		for _, k := range params.GoWords {
			val := 0.0

			keys := unigrams
			if strings.ContainsRune(k, '_') {
				keys = bigrams
			}

			for i := 0; i < len(keys); i++ {
				if strings.EqualFold(keys[i], k) {
					val = 1.0
					break
				}
			}

			features = append(features, val)
		}

		dateInfo := 0.0
		strTokens := strings.Split(
			strings.TrimSpace(
				sentSymbolsReplacePtr.ReplaceAllString(*sent.Text, " "),
			), " ")

		for _, s := range strTokens {
			isDate := false
			for _, dtPrt := range datePtr {
				if dtPrt.MatchString(s) {
					dateInfo = 1.0
					isDate = true
					break
				}
			}

			if isDate {
				break
			}
		}

		features = append(features, dateInfo)

		x := make([]svm.Node, len(features))
		for j := 0; j < len(features); j++ {
			x[j] = svm.Node{
				Index: j + 1,
				Value: features[j],
			}
		}

		clsLabel := svm.Predict(params.Model, x)

		clsVal := "UNKNOWN"
		switch clsLabel {
		case ClassCurrSmokerInt:
			clsVal = ClassCurrSmoker
		case ClassPastSmokerInt:
			clsVal = ClassPastSmoker
		case ClassSmokerInt:
			clsVal = ClassSmoker
		}

		return clsVal

	}, nil
}
