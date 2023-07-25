package main

import (
	"bufio"
	"text2phenotype.com/fdl/pipeline"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"
)

type SampleData struct {
	Name string
	Path string
}

func readSamples(inDir string) ([]*SampleData, error) {
	fInfos, err := ioutil.ReadDir(inDir)
	if err != nil {
		return nil, err
	}

	var data []*SampleData

	for _, fInfo := range fInfos {
		if !fInfo.IsDir() && strings.HasSuffix(fInfo.Name(), ".txt") {
			newSampleData := SampleData{
				Name: fInfo.Name(),
				Path: path.Join(inDir, fInfo.Name()),
			}
			data = append(data, &newSampleData)
		}
	}
	return data, nil
}

func TestBatchProcessing(t *testing.T) {

	// Folder with configurations: <fdl repository folder>/fdl/config
	cfgDir := ""
	// Folder with samples *.txt
	inDir := ""
	// Folder to save results *.json
	outDir := ""
	// Dictionaries folder: ,<fdl repository folder>/fdl/resources/dictionaries
	dictDir := ""
	// Resources folder: ,<fdl repository folder>/go/resources
	resPath := ""
	// Number of samples which will processed in parallel
	batchSize := 10

	cfgs, err := types.LoadConfigurations(cfgDir)
	if err != nil {
		t.Fatal(err)
	}

	params := pipeline.DefaultClinicalParams{
		DictionaryFolder: dictDir,
		Configurations:   cfgs,
		ResourceFolder:   resPath,
		LabValuesParams: pipeline.LabValuesParams{
			Model:            path.Join(resPath, "lab_values/model.json"),
			MaxTokenDistance: 15,
			LabUnitsFile:     path.Join(resPath, "lab_values/units.txt"),
			StringValues:     []string{"normal"},
		},
	}
	ppln, err := pipeline.DefaultClinical(params)
	if err != nil {
		t.Fatal(err)
	}

	utils.GlobalStringStore().Lock()

	sampleData, err := readSamples(inDir)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(sampleData))

	var batchGroup sync.WaitGroup

	samplesCh := make(chan *SampleData, batchSize)

	vocabParams := make(map[int]pipeline.RequestVocabParams)
	for i, cfg := range cfgs {
		vocabParams[i] = pipeline.RequestVocabParams{
			Name:   cfg.Name,
			Params: cfg.RequestParams,
		}
	}

	// processing
	go func() {

		for data := range samplesCh {
			buf, err := ioutil.ReadFile(data.Path)
			if err != nil {
				t.Error(err)
				wg.Done()
				return
			}

			txt := string(buf)

			req := pipeline.Request{
				Tid:  data.Name,
				Text: txt,
			}

			go func(r pipeline.Request, dt *SampleData) {
				defer wg.Done()
				defer batchGroup.Done()

				resp := <-ppln(r)
				outPath := path.Join(outDir, dt.Name+".json")
				f, err := os.Create(outPath)
				if err != nil {
					t.Fatal(err)
				}

				w := bufio.NewWriter(f)
				_, err = w.Write([]byte(resp))
				if err != nil {
					t.Fatal(err)
				}
				err = w.Flush()
				if err != nil {
					t.Fatal(err)
				}

			}(req, data)

		}
	}()

	t0 := time.Now()
	// send to process
	for i, data := range sampleData {
		if i%batchSize == 0 {
			batchGroup.Wait()
		}
		batchGroup.Add(1)
		samplesCh <- data
	}

	wg.Wait()

	println("Processing time", time.Since(t0).Milliseconds())
}
