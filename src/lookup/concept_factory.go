package lookup

import (
	"bufio"
	"bytes"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"io"
	"io/ioutil"
	"strings"
)

type SemanticConcepts map[types.Semantic]*types.Concept // semantic -> concept

type ConceptFactory func(cuis []*string) (map[*string]SemanticConcepts, error) // cui -> semantic concepts

type ConceptOffset struct {
	Offset int64
	Length int
}

func parseTUIs(tui string) []string {
	if strings.HasPrefix(tui, "[") {
		return strings.Split(tui[1:len(tui)-1], ",")
	} else {
		return []string{tui}
	}
}

func CreateConceptFactory(configName string, path string, scheme []string, ignoreParams []string) (ConceptFactory, error) {
	fdlLogger := logger.NewLogger("Concept factory loader").With().
		Str("config_name", configName).
		Str("path", path).Logger()
	fdlLogger.Info().Msg("Started loading")

	// create scheme index map
	var schemaMap = make(map[string]int)
	for i, columnName := range scheme {
		schemaMap[columnName] = i
	}

	// remove ignored columns from the scheme
	if len(ignoreParams) > 0 {
		for _, p := range ignoreParams {
			delete(schemaMap, p)
		}
	}

	cuiIdx := schemaMap[types.CUI]

	dictBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	bytesReader := bytes.NewReader(dictBytes)
	reader := bufio.NewReader(bytesReader)
	store := utils.GlobalStringStore()

	conceptMap := make(map[*string][]ConceptOffset)

	offset := int64(0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		offset += int64(len(line))
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		cutLine := strings.ToLower(strings.Trim(line, "\n"))
		columns := strings.Split(cutLine, "|")

		cui := columns[cuiIdx]
		cuiPtr := store.GetPointer(cui)

		conceptData := ConceptOffset{
			Offset: offset - int64(len(line)),
			Length: len(line),
		}
		conceptMap[cuiPtr] = append(conceptMap[cuiPtr], conceptData)
	}

	tuiIdx := schemaMap[types.TUI]

	fdlLogger.Info().Msgf("Loaded %d concepts", len(conceptMap))
	errLogger := fdlLogger.With().Caller().Logger()

	return func(cuis []*string) (map[*string]SemanticConcepts, error) {
		result := make(map[*string]SemanticConcepts)

		for _, cuiPtr := range cuis {
			cuiOffsets := conceptMap[cuiPtr]

			cuiSemantics := make(SemanticConcepts)

			for _, cuiOffset := range cuiOffsets {
				buf := make([]byte, cuiOffset.Length)
				_, err := bytesReader.ReadAt(buf, cuiOffset.Offset)
				if err != nil {
					errLogger.Err(err).Msg("")
					return nil, err
				}
				line := string(buf)

				line = strings.ToLower(strings.Trim(line, "\n"))
				columns := strings.Split(line, "|")

				tuis := parseTUIs(columns[tuiIdx])

				for _, tui := range tuis {
					tuiSemantic := GutTUISemanticGroupID(tui)
					semanticConcept, ok := cuiSemantics[tuiSemantic]
					if !ok {
						semanticConcept = types.CreateConcept(columns, cuiPtr, schemaMap)
					}
					semanticConcept.Update(tui, columns, schemaMap)
					cuiSemantics[tuiSemantic] = semanticConcept

				}
			}
			result[cuiPtr] = cuiSemantics
		}
		return result, nil
	}, nil
}
