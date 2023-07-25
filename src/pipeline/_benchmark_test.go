package pipeline

import (
	"text2phenotype.com/fdl/types"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type BenchmarkEnvConfig struct {
	ConfigDirPath            string `envconfig:"FDL_CONFIG_PATH" required:"true"`          // Should point to a directory with FDL configs
	FDLRootDirPath           string `envconfig:"FDL_DIR_PATH" required:"true"`             // Should point to FDL's root directory
	DictionaryPath           string `envconfig:"FDL_DICTIONARY_PATH" required:"true"`      // Should point to FDL dictionaries directory
	TextSamplesDirectoryPath string `envconfig:"TEXT_SAMPLES_PATH" required:"true"`        // Should point to a directory containing .txt files from text2phenotype-samples
	GoMaxProcesses           int    `envconfig:"GOMAXPROCS" required:"true"`               // Limit value set in FDL service configuration in yacht-anchor is 3
	MaxParallelRequestCount  int    `envconfig:"FDL_MQ_MAX_PARALLEL_REQUESTS" default:"5"` // This variable sets the number of text samples processed in parallel during benchmark
}

// This benchmark is designed to be used for pipeline profiling.
// It requires text samples in .txt format to run, so it's primarily intended to be run in local dev environments.
// To run, several environment variables listed in BenchmarkEnvConfig structure should be set.
// If number of text samples is significant, number of times each sample is processed can be set to N with -test.benchtime=Nx program argument
func BenchmarkPipelineOnTextSamples(b *testing.B) {
	var config BenchmarkEnvConfig
	err := envconfig.Process("", &config)
	if err != nil {
		b.Fatal(err)
	}
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	cfgs, err := types.LoadConfigurations(config.ConfigDirPath)
	if err != nil {
		b.Fatal(err)
	}
	params := GetDefaultClinicalParams(config.FDLRootDirPath, config.DictionaryPath, cfgs)
	ppln, err := DefaultClinical(params)
	if err != nil {
		b.Fatal(err)
	}
	workersCount := config.MaxParallelRequestCount

	preparedTasks := make([]Request, 0, 500)
	samplesPath := resolvePath(b, config.TextSamplesDirectoryPath)

	err = filepath.WalkDir(samplesPath, func(path string, d fs.DirEntry, err1 error) error {
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".txt") {
			return nil
		}
		text, err := loadText(filepath.Join(samplesPath, name))
		if err != nil {
			b.Fatal(err, name)
		}
		preparedTasks = append(preparedTasks, Request{
			Tid:  name,
			Text: text,
		})
		return nil
	})
	if err != nil {
		b.Fatal(err)
	}
	var wg sync.WaitGroup
	tasks := make(chan Request, workersCount)

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				for i := 0; i < b.N; i++ {
					<-ppln(task)
				}
			}
		}()
	}
	for _, task := range preparedTasks {
		tasks <- task
	}
	close(tasks)
	wg.Wait()
}

func resolvePath(b *testing.B, basePath string) string {
	dirInfo, err := os.Lstat(basePath)
	if err != nil {
		b.Fatal(err)
	}
	if dirInfo.Mode()&os.ModeSymlink == 0 {
		return basePath
	}
	readlink, err := os.Readlink(basePath)
	if err != nil {
		b.Fatal(err)
	}
	return readlink
}

func loadText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
