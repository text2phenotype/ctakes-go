package pipeline

import (
	"text2phenotype.com/fdl/logger"
	. "text2phenotype.com/fdl/lookup"
	"text2phenotype.com/fdl/types"
	"errors"
	"github.com/rs/zerolog"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

type DictionaryLookupParams struct {
	MinimumLookupSpan uint32
	ExclusionTags     []string
}

func GetDefaultDictionaryLookupParams() DictionaryLookupParams {
	return DictionaryLookupParams{
		MinimumLookupSpan: 1,
		ExclusionTags:     []string{"VB"},
	}
}

type LookupConfig struct {
	Name    string
	Dict    Dictionary
	Factory ConceptFactory
	Cons    Consumer
	Params  DictionaryLookupParams
}

func CreateLookupConfigs(dictDir string, configs []types.Configuration) (map[string]LookupConfig, error) {

	fdlLogger := logger.NewLogger("CreateLookupConfigs")
	fdlLogger.Info().Msg("Loading configurations")

	var wg sync.WaitGroup

	dictMap := make(map[string]Dictionary)
	factoryMap := make(map[string]ConceptFactory)

	for _, cfg := range configs {
		configLogger := fdlLogger.With().Str("config_name", cfg.Name).Logger()

		if cfg.Pipeline != types.DefaultClinicalPipeline {
			continue
		}
		wg.Add(1)
		// load dictionary
		errLogger := configLogger.With().Caller().Logger()
		go func(configName string, dictParams types.FDLConfig, errLogger zerolog.Logger) {
			defer wg.Done()

			dictPath := dictParams.TermDictionary
			if len(dictPath) == 0 {
				errLogger.Error().Str("path", dictPath).Msg("Dictionary path is not correct")
				return
			}

			absDictPath := path.Join(
				dictDir,
				dictPath,
			)

			dictSchemeParam := dictParams.TermScheme
			if len(dictSchemeParam) == 0 {
				errLogger.Error().Str("scheme", dictSchemeParam).Msg("Dictionary scheme is not correct")
				return
			}

			dictScheme := strings.Split(dictSchemeParam, "|")

			dict, e := CreateDictionary(configName, absDictPath, dictScheme)
			if e != nil {
				errLogger.Err(e).Msg("Could not load dictionary")
				return
			}

			dictMap[configName] = dict

		}(cfg.Name, cfg.Params.FDL, errLogger)

		wg.Add(1)
		// load concept factory
		go func(configName string, factParams types.FDLConfig, errLogger zerolog.Logger) {
			defer wg.Done()

			factPath := factParams.ConceptDictionary
			if len(factPath) == 0 {
				errLogger.Error().Str("factory_path", factPath).Msg("Concept factory path is not correct")
				return
			}

			absFactPath := path.Join(
				dictDir,
				factPath,
			)

			factSchemeParam := factParams.ConceptScheme
			if len(factSchemeParam) == 0 {
				errLogger.Error().Str("factory_scheme", factSchemeParam).Msg("Concept factory scheme is not correct")
				return
			}

			factScheme := strings.Split(factSchemeParam, "|")

			ignoredParams := factParams.ConceptIgnoredParams

			fact, e := CreateConceptFactory(configName, absFactPath, factScheme, ignoredParams)
			if e != nil {
				errLogger.Err(e).Msg("Could not create concept factory")
				return
			}
			factoryMap[factPath] = fact

		}(cfg.Name, cfg.Params.FDL, errLogger)
	}

	wg.Wait()

	result := make(map[string]LookupConfig)
	for _, cfg := range configs {

		dict, ok := dictMap[cfg.Name]
		if !ok {
			continue
		}
		factory, ok := factoryMap[cfg.Params.FDL.ConceptDictionary]
		if !ok {
			continue
		}
		if cfg.Pipeline == types.DefaultClinicalPipeline {
			lookupCfg := LookupConfig{
				Name:    cfg.Name,
				Dict:    dict,
				Factory: factory,
				Cons:    CreateConsumer(cfg.Params.FDL.PrecisionMode),
				Params:  GetDefaultDictionaryLookupParams(),
			}

			lookupCfg.Params.ExclusionTags = cfg.Params.FDL.ExclusionTags
			result[cfg.Name] = lookupCfg
		}
	}
	if len(result) == 0 {
		return nil, errors.New("failed to load at least one correct config")
	}
	fdlLogger.Info().Msgf("Loaded %d lookup configurations", len(result))
	return result, nil
}

func NewDictionaryLookup() func(in <-chan types.Sentence, cfg LookupConfig, tid string) <-chan []types.Annotation {
	return func(in <-chan types.Sentence, cfg LookupConfig, tid string) <-chan []types.Annotation {
		fdlLogger := logger.NewLogger("DictionaryLookup").With().
			Str("config_name", cfg.Name).
			Str("tid", tid).Logger()

		out := make(chan []types.Annotation)
		dictionary := cfg.Dict
		factory := cfg.Factory
		consumer := cfg.Cons

		go func() {
			defer close(out)
			var cnt uint32

			var wg sync.WaitGroup
			for sent := range in {

				wg.Add(1)

				go func(sent types.Sentence) {
					defer wg.Done()

					spans, cuis := searchSpansInWindow(sent, dictionary, cfg.Params)

					var allCuis []*string
					cuisCache := make(map[*string]bool)
					for _, spanCuis := range cuis {
						for _, cui := range spanCuis {
							_, hasCui := cuisCache[cui]
							if !hasCui {
								cuisCache[cui] = true
								allCuis = append(allCuis, cui)
							}
						}

					}
					conceptMap, err := factory(allCuis)
					if err != nil {
						fdlLogger.Err(err).Msg("")
						return
					}

					annotations := consumer(spans, cuis, conceptMap)
					for i := 0; i < len(annotations); i++ {
						annotations[i].Sentence = &sent
					}
					//for _, ann := range annotations {
					//	ann.Sentence = &sent
					//}

					atomic.AddUint32(&cnt, uint32(len(annotations)))

					// sort annotations
					sort.SliceStable(annotations, func(i, j int) bool {
						return annotations[i].Begin <= annotations[j].Begin
					})

					out <- annotations
				}(sent)

			}

			wg.Wait()
			fdlLogger.Debug().Msgf("Found %d annotations", cnt)
		}()
		return out
	}
}

func searchSpansInWindow(sentence types.Sentence, dictionary Dictionary, params DictionaryLookupParams) ([]types.Span, [][]*string) {

	spansMap := make(map[uint64]types.Span)
	cuiMap := make(map[uint64]map[*string]bool)

	nonNewLineIndices := getNonNewLineIndices(sentence)

	sentenceTextRunes := []rune(*sentence.Text)

	tokens := sentence.Tokens

	for idx, lookupIndex := range nonNewLineIndices {
		lookupToken := tokens[lookupIndex]
		if isNonLookupToken(*lookupToken, params) {
			continue
		}

		words := make([]*string, 0, 2)
		words = append(words, lookupToken.Text)

		if lookupToken.Lemma != nil && lookupToken.Text != lookupToken.Lemma {
			words = append(words, lookupToken.Lemma)
		}
		itr := dictionary(words)

		for {
			rareWordHit, ok := itr()
			if !ok {
				break
			}

			if rareWordHit.TextLength < params.MinimumLookupSpan {
				continue
			}

			if len(rareWordHit.Tokens) == 1 {
				spanHash := lookupToken.Span.GetHashCode()
				spanCuis, hasSpan := cuiMap[spanHash]
				if !hasSpan {
					spansMap[spanHash] = lookupToken.Span
					spanCuis = make(map[*string]bool)
				}
				spanCuis[rareWordHit.CUI] = true
				cuiMap[spanHash] = spanCuis
				continue
			}

			//termStartIndex := lookupIndex - int(rareWordHit.RareWordIndex)
			//if termStartIndex < 0 || termStartIndex+rareWordHit.GetTokenCount() > len(tokens) {
			//	continue
			//}

			lookupStartIndex := idx - int(rareWordHit.RareWordIndex)
			if lookupStartIndex < 0 || lookupStartIndex+rareWordHit.GetTokenCount() > len(nonNewLineIndices) {
				continue
			}

			termStartIndex := nonNewLineIndices[lookupStartIndex]

			//termEndIndex := termStartIndex + rareWordHit.GetTokenCount() - 1
			lookupEndIndex := lookupStartIndex + rareWordHit.GetTokenCount() - 1
			if lookupEndIndex >= len(nonNewLineIndices) {
				continue
			}

			termEndIndex := nonNewLineIndices[lookupEndIndex]

			if isTermMatch(rareWordHit, tokens, termStartIndex, termEndIndex) {
				spanStart := tokens[termStartIndex].Begin
				spanEnd := tokens[termEndIndex].End
				//spanText := (*sentence.Text)[spanStart-sentence.Begin : spanEnd-sentence.Begin]
				spanText := string(sentenceTextRunes[spanStart-sentence.Begin : spanEnd-sentence.Begin])
				newSpan := types.Span{
					Begin: spanStart,
					End:   spanEnd,
					Text:  &spanText,
				}

				spanHash := newSpan.GetHashCode()
				spanCuis, hasSpan := cuiMap[spanHash]
				if !hasSpan {
					spansMap[spanHash] = newSpan
					spanCuis = make(map[*string]bool)
				}
				spanCuis[rareWordHit.CUI] = true
				cuiMap[spanHash] = spanCuis
			}
		}
	}

	spans := make([]types.Span, 0, len(spansMap))
	cuis := make([][]*string, 0, len(spansMap))

	for h, span := range spansMap {
		spans = append(spans, span)

		cuiToSlice := make([]*string, len(cuiMap[h]))
		i := 0
		for cui := range cuiMap[h] {
			cuiToSlice[i] = cui
			i++
		}

		cuis = append(cuis, cuiToSlice)
	}
	return spans, cuis
}

func getNonNewLineIndices(sentence types.Sentence) []int {
	var result []int
	for i, token := range sentence.Tokens {
		if !token.IsNewline {
			result = append(result, i)
		}
	}
	return result
}

func isTermMatch(term *RareWordTerm, tokens []*types.Token, beginIdx int, endIdx int) bool {
	hitTokens := term.Tokens
	hit := 0
	for i := beginIdx; i < endIdx+1; i++ {
		if tokens[i].IsNewline {
			continue
		}
		if hitTokens[hit] == tokens[i].Lemma || hitTokens[hit] == tokens[i].Text {
			hit++
			continue
		}

		return false
	}
	return true
}

func isNonLookupToken(token types.Token, params DictionaryLookupParams) bool {
	toExclude := false
	for _, tag := range params.ExclusionTags {
		if strings.EqualFold(tag, *token.Tag) {
			toExclude = true
			break
		}
	}

	return token.IsNewline || token.IsPunct || toExclude
}
