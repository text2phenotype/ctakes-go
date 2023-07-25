package main

import (
	"text2phenotype.com/fdl/api"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/pipeline"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"text2phenotype.com/fdl/worker"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"net/http"
	"os"
	"time"
)

type Config struct {
	ConfigPath     string `envconfig:"FDL_CONFIG_PATH" required:"true"`
	DictionaryPath string `envconfig:"FDL_DICTIONARY_PATH" required:"true"`
	DirPath        string `envconfig:"FDL_DIR_PATH" required:"true"`
	RestAPIActive  bool   `envconfig:"FDL_REST_API_ACTIVE" default:"false"`
	RestAPIPort    string `envconfig:"FDL_REST_API_PORT" default:"10000"`
}

const pipelineStartMaxRetries = 5

func main() {
	logger.SetupLogging()
	fdlLogger := logger.NewLogger("Main")
	fatalErrLogger := fdlLogger.Fatal().Caller()
	buildIndex := flag.Bool("build-index", false, "a bool")
	flag.Parse()
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		fatalErrLogger.Err(err).Msg("Failed to read environment")
		os.Exit(1)
	}

	// build config index
	if *buildIndex {
		cfgs, err := types.LoadConfigurations(config.ConfigPath)
		if err != nil {
			fdlLogger.Err(err).Msg("Failed to load configurations")
			return
		}
		_, err = pipeline.CreateLookupConfigs(config.DictionaryPath, cfgs)
		if err != nil {
			fatalErrLogger.Err(err).Msg("Failed to build config indexes cache")
			os.Exit(1)
		} else {
			fdlLogger.Info().Msg("Configs indexes cache was built. Exit...")
		}
		return
	}

	//Load Pipeline
	pipelineChannel := make(chan pipeline.Pipeline)
	go func() {
		for retry := 0; retry < pipelineStartMaxRetries; retry++ {
			cfgs, err := types.LoadConfigurations(config.ConfigPath)
			if err != nil {
				fdlLogger.Err(err).Msg("Failed to load configurations. Retrying in 5 sec")
				time.Sleep(5 * time.Second)
				continue
			}
			fdlLogger.Info().Msgf("Loaded %d configurations", len(cfgs))
			fdlLogger.Info().Msg("Starting pipelines loading")

			pipelineParams := pipeline.GetDefaultClinicalParams(config.DirPath, config.DictionaryPath, cfgs)
			ppln, err := pipeline.DefaultClinical(pipelineParams)
			if err != nil {
				fdlLogger.Err(err).Msg("Failed to start default clinical pipeline. Retrying in 5 sec")
				time.Sleep(5 * time.Second)
				continue
			}
			utils.GlobalStringStore().Lock()
			fdlLogger.Info().Msg("Pipelines loaded")
			pipelineChannel <- ppln
			return
		}
		fatalErrLogger.Msg("Could not start pipelines after 5 retries, exiting")
		os.Exit(1)
	}()

	// block until pipeline loads
	ppln := <-pipelineChannel

	if config.RestAPIActive {
		go func() {
			fdlLogger.Info().Msg("Starting API service")
			apiRequest := &api.Request{
				Pipeline: ppln,
			}
			http.HandleFunc("/", apiRequest.ProcessData)
			host := fmt.Sprintf(":%s", config.RestAPIPort)
			fdlLogger.Info().Msgf("REST API on %s", host)
			err := http.ListenAndServe(host, nil)
			fatalErrLogger.Err(err).Msg("REST API stopped with error")
		}()
	}

	fdlLogger.Info().Msg("Start FDL Worker")
	for {
		rmqWorker, err := worker.New(ppln)
		if err != nil {
			fdlLogger.Fatal().Err(err).Msg("Could not initialize RMQ worker")
			os.Exit(1)
		}
		err = rmqWorker.StartWorker()
		if err != nil {
			fdlLogger.Err(err).Msg("Worker returned with error. Launching new in 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}
}
