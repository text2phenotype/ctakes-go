package pipeline

import (
	drugFsm "text2phenotype.com/fdl/drug_ner/fsm"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/negation"
	"text2phenotype.com/fdl/nlp"
	"text2phenotype.com/fdl/pos"
	"text2phenotype.com/fdl/types"
	"encoding/json"
	"path"
)

type DrugNerParams struct {
	MaxAttributeDistance int `json:"max_attribute_distance"`
}

type LabValuesParams struct {
	StringValues     []string `json:"string_values"`
	MaxTokenDistance int      `json:"max_token_distance"`
	Model            string   `json:"model"`
	LabUnitsFile     string   `json:"lab_units_file"`
}

type DefaultClinicalParams struct {
	DictionaryFolder string                `json:"dictionary_folder"`
	ResourceFolder   string                `json:"resource_folder"`
	Configurations   []types.Configuration `json:"configurations"`
	LabValuesParams  LabValuesParams       `json:"lab_values_params"`
	DrugParams       DrugNerParams         `json:"drug_params"`
}

func GetDefaultClinicalParams(filePath string, dictPath string, cfgs []types.Configuration) DefaultClinicalParams {
	resourcesPath := path.Join(filePath, "resources")
	return DefaultClinicalParams{
		DictionaryFolder: dictPath,
		Configurations:   cfgs,
		ResourceFolder:   resourcesPath,
		LabValuesParams: LabValuesParams{
			Model:            path.Join(resourcesPath, "lab_values", "model.json"),
			MaxTokenDistance: 15,
			LabUnitsFile:     path.Join(resourcesPath, "lab_values", "units.txt"),
			StringValues:     []string{"normal"},
		},
	}
}

func DefaultClinical(params DefaultClinicalParams) (Pipeline, error) {
	fdlLogger := logger.NewLogger("Default Clinical pipeline")
	errLogger := fdlLogger.With().Caller().Logger()
	fdlLogger.Info().
		Interface("params", params).
		Msg("Starting default clinical pipeline (see parameters in 'params' field)")
	lookupCfg, err := CreateLookupConfigs(params.DictionaryFolder, params.Configurations)
	if err != nil {
		errLogger.Err(err).
			Interface("configurations", params.Configurations).
			Str("dictionary_folder", params.DictionaryFolder).
			Msg("Failed to create lookup config")
		return nil, err
	}

	sentDetectorResources := path.Join(params.ResourceFolder, "sentdetector")
	sentenceDetector, err := nlp.NewSentenceDetector(sentDetectorResources)
	if err != nil {
		errLogger.Err(err).
			Str("sent_detector_resources_path", sentDetectorResources).
			Msg("Failed to create sentence detector")
		return nil, err
	}

	tokenizer, err := NewTokenizer()
	if err != nil {
		errLogger.Err(err).Msg("Failed to create tokenizer")
		return nil, err
	}

	posModelLocation := path.Join(params.ResourceFolder, "pos", "pos.model.json")
	posModel, err := pos.LoadModelFromFile(posModelLocation)
	if err != nil {
		errLogger.Err(err).
			Str("pos_model_location", posModelLocation).
			Msg("Failed to load POS model")
		return nil, err
	}

	tagger := NewPOSTagger(posModel)

	lemmatizerResources := path.Join(params.ResourceFolder, "lemmatizer")
	lemmatizer, err := NewLemmatizer(lemmatizerResources)
	if err != nil {
		errLogger.Err(err).
			Str("lemmatizer_resources_folder", lemmatizerResources).
			Msg("Failed to create lemmatizer")
		return nil, err
	}

	lookup := NewDictionaryLookup()
	splitter := NewSentenceChannelSplitter(len(params.Configurations))

	labAttributesAnnotator, err := NewLabAttributesAnnotator(params.LabValuesParams)
	if err != nil {
		errLogger.Err(err).
			Interface("lab_values_params", params.LabValuesParams).
			Msg("Failed to create lab attributes annotator")
		return nil, err
	}

	drugFSMData, err := drugFsm.LoadDrugFSMExtractorParams(params.ResourceFolder)
	if err != nil {
		errLogger.Err(err).
			Str("resource_folder", params.ResourceFolder).
			Msg("Failed to load drug fsm data")
		return nil, err
	}

	drugAttributesAnnotator, err := NewDrugAttributesAnnotator(params.DrugParams, drugFSMData)
	if err != nil {
		errLogger.Err(err).
			Interface("drug_params", params.DrugParams).
			Msg("Failed to create drug attributes annotator")
		return nil, err
	}

	defaultBoundaries := negation.GetDefaultBoundaries()
	polarityDetectorParams := struct {
		MaxLeftScopeSize  int           `json:"max_left_scope_size"`
		MaxRightScopeSize int           `json:"max_right_scope_size"`
		Scopes            []types.Scope `json:"scopes"`
	}{20, 10, []types.Scope{types.ScopeLeft, types.ScopeRight}}

	polarityDetector, err := NewPolarityDetector(
		polarityDetectorParams.MaxLeftScopeSize,
		polarityDetectorParams.MaxRightScopeSize,
		polarityDetectorParams.Scopes,
		defaultBoundaries)

	if err != nil {
		errLogger.Err(err).
			Interface("polarity_detector_params", polarityDetectorParams).
			Interface("default_boundaries", defaultBoundaries).
			Msg("Failed to create polarity detector")
		return nil, err
	}

	adjusterParams := SentenceAdjusterParams{
		wordsInPattern: []string{"no", "none", "never", "quit", "smoked", ":"},
	}
	sentenceAdjuster := NewSentenceAdjuster(adjusterParams)

	smokingResources := path.Join(params.ResourceFolder, "smoking")
	smokingStatusParams, err := LoadSmokingStatusParameters(smokingResources)
	if err != nil {
		errLogger.Err(err).
			Str("smoking_resources_folder", smokingResources).
			Msg("Failed to create smoking status parameters")
		return nil, err
	}

	smokingStatusAnnotator, err := NewSmokingStatusAttributeAnnotator(smokingStatusParams)
	if err != nil {
		errLogger.Err(err).
			Interface("smoking_status_params", smokingStatusParams).
			Msg("Failed to create smoking status annotator")
		return nil, err
	}

	smoking_response := NewSmokingStatusResult()

	default_response := NewDefaultClinicalResult()

	return func(request Request) <-chan string {
		responseChan := make(chan string)
		pplnLog := fdlLogger.With().Str("tid", request.Tid).Logger()
		pplnLog.Info().Msg("Started default clinical pipeline")
		errLogger = pplnLog.With().Caller().Logger()

		go func() {
			var in = make(chan string)

			sd := sentenceDetector(in)
			tok := tokenizer(sd)
			tag := tagger(tok)
			lem := lemmatizer(tag)

			split := splitter(lem)

			resultChannel := make(chan Result)
			defer close(resultChannel)

			for i, cfg := range params.Configurations {
				switch cfg.Pipeline {
				case types.DefaultClinicalPipeline:
					{
						lCfg := lookupCfg[cfg.Name]

						annotations := lookup(split[i], lCfg, request.Tid)

						if cfg.CheckFeature(types.LabAttributes) {
							annotations = labAttributesAnnotator(annotations)
						}

						if cfg.CheckFeature(types.DrugAttributes) {
							annotations = drugAttributesAnnotator(annotations)
						}
						if cfg.CheckFeature(types.PolarityAttributes) {
							annotations = polarityDetector(annotations)
						}

						defRes := default_response(annotations, lCfg, request)
						connect(defRes, resultChannel)

					}
				case types.SmokingStatusPipeline:
					{
						annotations := sentenceAdjuster(split[i])
						annotations = smokingStatusAnnotator(annotations)
						smokRes := smoking_response(annotations, cfg.Name, request)
						connect(smokRes, resultChannel)
					}
				}
			}

			in <- request.Text
			close(in)
			response := make(map[string]interface{})

			for i := 0; i < len(params.Configurations); i++ {
				res := <-resultChannel
				pplnLog.Info().
					Str("config_name", res.ConfigName).
					Msg("Finished pipeline for configuration")
				response[res.ConfigName] = res.Data
			}

			buf, err := json.Marshal(response)
			if err != nil {
				errLogger.Err(err).Str("tid", request.Tid).Msg("Failed to marshall response")
			}
			pplnLog.Info().Msg("Finished default clinical pipeline")
			txt := string(buf)
			responseChan <- txt
		}()

		return responseChan
	}, nil

}

func connect(from <-chan Result, to chan<- Result) {
	go func() {
		for v := range from {
			to <- v
		}
	}()
}
