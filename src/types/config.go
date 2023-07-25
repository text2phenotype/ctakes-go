package types

import (
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/utils"
	"errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	LookupModeDefault = "str"
	LookupModeCode    = "code"

	// pipeline type
	DefaultClinicalPipeline = "default_clinical"
	SmokingStatusPipeline   = "smoking_status"

	// features
	LabAttributes      = "lab"
	DrugAttributes     = "drug"
	PolarityAttributes = "polarity"
)

type RequestParams struct {
	LookupMode string `yaml:"lookup_mode" json:"lookup_mode"`
}

func (rParams RequestParams) IsEmpty() bool {
	return len(rParams.LookupMode) == 0
}

func (rParams RequestParams) GetHashCode() uint64 {
	if rParams.LookupMode == "" {
		rParams.LookupMode = LookupModeDefault
	}
	return utils.HashString(strings.ToLower(rParams.LookupMode))
}

type FDLConfig struct {
	TermDictionary       string   `yaml:"term_dictionary" json:"term_dictionary"`
	TermScheme           string   `yaml:"term_scheme" json:"term_scheme"`
	ConceptDictionary    string   `yaml:"concept_dictionary" json:"concept_dictionary"`
	ConceptScheme        string   `yaml:"concept_scheme" json:"concept_scheme"`
	ConceptIgnoredParams []string `yaml:"concept_params_ignore" json:"concept_ignored_params"`
	ExclusionTags        []string `yaml:"exclusion_tags" json:"exclusion_tags"`
	PrecisionMode        bool     `yaml:"precision_mode" json:"precision_mode"`
}

type ParamsConfig struct {
	FDL FDLConfig `yaml:"FDL" json:"fdl"`
}
type Configuration struct {
	Name          string        `json:"name"`
	FilePath      string        `json:"file_path"`
	RequestParams RequestParams `yaml:"request_params" json:"request_params"`
	Params        ParamsConfig  `yaml:"params" json:"params"`
	Pipeline      string        `yaml:"pipeline" json:"pipeline"`
	Features      []string      `yaml:"features" json:"features"`
}

func (cfg Configuration) CheckFeature(featureName string) bool {
	for _, feat := range cfg.Features {
		if feat == featureName {
			return true
		}
	}

	return false
}

func LoadConfigurations(dirPath string) ([]Configuration, error) {
	fdlLogger := logger.NewLogger("LoadConfigurations")

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	configChan := make(chan Configuration, len(files))
	for _, f := range files {
		// Skip dirs and non-yaml files
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".yaml") {
			continue
		}

		wg.Add(1)
		go func(file os.FileInfo) {
			defer wg.Done()
			cfg := Configuration{
				Name:     strings.Split(file.Name(), ".yaml")[0],
				FilePath: path.Join(dirPath, file.Name()),
			}
			buf, err := ioutil.ReadFile(cfg.FilePath)
			if err != nil {
				fdlLogger.Err(err)
				return
			}
			if err := yaml.Unmarshal(buf, &cfg); err != nil {
				fdlLogger.Err(err)
				return
			}

			// check pipeline type
			if cfg.Pipeline != DefaultClinicalPipeline && cfg.Pipeline != SmokingStatusPipeline {
				fdlLogger.Err(errors.New("wrong pipeline type"))
				return
			}

			configChan <- cfg
		}(f)
	}

	go func() {
		wg.Wait()
		close(configChan)
	}()

	// TODO: in the future channel with configs can send configs into pipeline directly
	configs := make([]Configuration, 0, len(configChan))
	for cfg := range configChan {
		configs = append(configs, cfg)
	}
	return configs, nil
}
