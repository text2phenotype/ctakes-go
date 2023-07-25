package pipeline

import (
	"archive/zip"
	"text2phenotype.com/fdl/types"
	"encoding/json"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func BenchmarkPipeline(b *testing.B) {

	tData := NewTestdata()
	for ctakesRes := range tData.CtakesResultCh {
		for i := 0; i < b.N; i++ {
			req := Request{
				Text: ctakesRes.Text,
				Tid:  ctakesRes.Filename,
			}
			<-tData.Pipeline(req)
		}
		// Only one sample
		//nolint
		break
	}
}

func TestCTakesComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped test for comparison output")
	}
	tData := NewTestdata()
	_ = os.RemoveAll(tData.ReportPath)

	annCounter := 0
	failCounter := 0
	diff := 0
	target := 1
	for ctakesRes := range tData.CtakesResultCh {
		fmt.Printf("Sample name: %s\n\n", ctakesRes.Filename)

		req := Request{
			Text: ctakesRes.Text,
			Tid:  ctakesRes.Filename,
		}
		res := <-tData.Pipeline(req)

		response := make(map[string]TestResponse)
		if err := json.Unmarshal([]byte(res), &response); err != nil {
			t.Fatal("Unmarshal failed, ", err)
		}

		for _, ctakes := range ctakesRes.Results {
			fmt.Printf("Config name: %s\n", ctakes.Config)

			annotation, ok := response[ctakes.Config]
			if !ok {
				t.Logf("FDL response does not have results for config - \"%s\"", ctakes.Config)
				continue
			}

			if ctakes.Response.Content == nil {
				ctakes.Response.Content = ctakes.Response.DrugEntities
			}
			if ctakes.Response.Content == nil {
				ctakes.Response.Content = ctakes.Response.LabValues
			}

			sort.Sort(annotation.Content)
			sort.Sort(ctakes.Response.Content)
			differenceCh := ctakes.Response.Diff(annotation)

			result := make([]Difference, 0)
			indexSet := make(map[int]bool)
			for item := range differenceCh {
				indexSet[item.id] = true
				item.Sentence = req.Text[item.orig.Sentence[0]:item.orig.Sentence[1]]
				item.Token = item.orig.Span[0].(string)
				result = append(result, item)
			}
			failCounter += len(indexSet)
			annCounter += len(ctakes.Response.Content)

			// Reports
			results := []Results{
				{
					filename: "ctakes.json",
					data:     ctakes.Response,
				},
				{
					filename: "fdl.json",
					data:     annotation,
				},
			}
			if len(result) != 0 {
				results = append(results, Results{
					filename: "errors.json",
					data:     result,
				})
			}
			for _, item := range results {
				tData.SaveResults(ctakesRes.Filename, ctakes.Config, item)
			}
		}
		// Only one sample
		break
	}

	diff = failCounter * 100 / annCounter
	require.LessOrEqual(t, diff, target, "Diff, %: ", diff)
}

func (r TestResponse) Diff(response TestResponse) <-chan Difference {
	difference := make(chan Difference)
	go func() {
		defer close(difference)

		var index int
		check := func(orig TestAnnotation, expected, received interface{}, reason string) {

			var trans []cmp.Option
			if cmp.Equal(expected, received, trans...) {
				return
			}

			difference <- Difference{
				id:       index,
				orig:     orig,
				Expected: expected,
				Received: received,
				Reason:   reason,
			}
		}

		q := r.Content.getMapAnnotations()
		ccc := response.Content.getMapAnnotations()
		for key, value := range q {
			c, ok := ccc[key]
			if !ok {
				check(value[0], nil, nil, "Token not found in FDL results")
				index++
				continue
			}

			for _, i := range value {
				isFound := false
				for _, j := range c {
					if i.Name != j.Name {
						continue
					}
					index++
					isFound = true

					check(i, i.Attributes, j.Attributes, "Attributes are different")

					for _, originConcept := range i.UmlsConcepts {
						for _, compConcept := range j.UmlsConcepts {
							if !strings.EqualFold(originConcept.Cui, compConcept.Cui) {
								continue
							}

							diff := originConcept.GetDiff(compConcept)
							for _, d := range diff {
								check(i, d.Expected, d.Received, d.Reason)
							}
						}
					}
					break
				}
				if !isFound {
					check(i, i, nil, "Annotation not found")
				}
			}
		}
	}()
	return difference
}

func (t Testdata) SaveResults(filename, config string, save Results) {
	configReportPath := path.Join(
		t.ReportPath,
		"comparison",
		strings.TrimSuffix(filename, filepath.Ext(filename)),
		config,
	)
	_ = os.MkdirAll(configReportPath, os.ModePerm)

	filePath := path.Join(configReportPath, save.filename)
	bytes, err := json.MarshalIndent(save.data, "", "\t")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filePath, bytes, 0644)
	if err != nil {
		return
	}
}

type testAnnotationSlice []TestAnnotation

func (t testAnnotationSlice) getMapAnnotations() map[string][]TestAnnotation {
	m := make(map[string][]TestAnnotation)

	for _, item := range t {
		key := fmt.Sprintf("%s_%d_%d",
			strings.ToLower(item.Span[0].(string)), int(item.Span[1].(float64)), int(item.Span[2].(float64)))
		m[key] = append(m[key], item)
	}
	return m
}

func (t testAnnotationSlice) Len() int      { return len(t) }
func (t testAnnotationSlice) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t testAnnotationSlice) Less(i, j int) bool {
	if cmp.Equal(t[i].Span, t[j].Span) {
		return t[i].Aspect > t[j].Aspect
	}
	startI := t[i].Span[1].(float64)
	stopI := t[i].Span[2].(float64)
	startJ := t[j].Span[1].(float64)
	stopJ := t[j].Span[2].(float64)

	switch {
	case startI < startJ:
		return stopI < stopJ
	case startI == startJ:
		return stopI < stopJ
	}
	return false
}

type AnnotationResult struct {
	Config   string
	Response TestResponse
}

type CtakesFileResult struct {
	Filename string
	Text     string
	Results  []AnnotationResult
}

type TestResponse struct {
	Jira          string              `json:"jira,omitempty"`
	User          string              `json:"user,omitempty"`
	Timestamp     string              `json:"timestamp,omitempty"`
	Date          string              `json:"date,omitempty"`
	Version       string              `json:"version,omitempty"`
	DocId         string              `json:"docId,omitempty"`
	Dob           string              `json:"dob,omitempty"`
	Gender        string              `json:"gender,omitempty"`
	Age           string              `json:"age,omitempty"`
	SmokingStatus string              `json:"smokingStatus,omitempty"`
	Sentences     []smokeItem         `json:"sentences,omitempty"`
	Content       testAnnotationSlice `json:"content,omitempty"`
	DrugEntities  testAnnotationSlice `json:"drugEntities,omitempty"`
	LabValues     testAnnotationSlice `json:"labValues,omitempty"`
}

type smokeItem struct {
	Status string `json:"status"`
	Test   Span   `json:"Test"`
}

type Difference struct {
	id       int
	orig     TestAnnotation
	Sentence string
	Token    string
	Expected interface{}
	Received interface{}
	Reason   string
}

type TestAnnotation struct {
	Attributes    `json:"attributes"`
	Span          `json:"Text"`
	Sentence      []int         `json:"Sentence"`
	SectionOffset []int         `json:"sectionOffset"`
	SectionOid    string        `json:"sectionOid"`
	Aspect        string        `json:"aspect"`
	Name          string        `json:"name"`
	UmlsConcepts  []UmlsConcept `json:"umlsConcepts"`
}

type UmlsConcept struct {
	Tui           []string     `json:"tui,omitempty"`
	Cui           string       `json:"cui,omitempty"`
	PreferredText string       `json:"preferredText,omitempty"`
	SabConcepts   []SabConcept `json:"sabConcepts,omitempty"`
}

func (origin UmlsConcept) GetDiff(comp UmlsConcept) []Difference {
	if origin.Cui != comp.Cui {
		return nil
	}
	diff := make([]Difference, 0)

	tuiSet := make(map[string]bool)
	for _, tui := range comp.Tui {
		tuiSet[strings.ToLower(tui)] = true
	}
	for _, tui := range origin.Tui {
		if !tuiSet[strings.ToLower(tui)] {
			diff = append(diff, Difference{
				Reason:   "Such TUI was not found",
				Expected: origin.Tui,
				Received: comp.Tui,
			})
		}
	}

	codingSchemeSet := make(map[string]SabConcept)
	for _, subConcept := range origin.SabConcepts {
		codingSchemeSet[strings.ToLower(subConcept.CodingScheme)] = subConcept
	}

	for _, subConcept := range comp.SabConcepts {
		origSabConcept, ok := codingSchemeSet[strings.ToLower(subConcept.CodingScheme)]
		if !ok {
			diff = append(diff, Difference{
				Reason:   "Such codingScheme was not found",
				Expected: subConcept.CodingScheme,
			})
			continue
		}

		ttySet := make(map[string]bool)
		codeSet := make(map[string]bool)
		for _, vocab := range subConcept.VocabConcepts {
			for _, tty := range vocab.Tty {
				ttySet[strings.ToLower(tty)] = true
			}
			codeSet[strings.ToLower(vocab.Code)] = true
		}

		for _, vocab := range origSabConcept.VocabConcepts {
			for _, tty := range vocab.Tty {
				if !ttySet[strings.ToLower(tty)] {
					diff = append(diff, Difference{
						Reason:   "Such TTY was not found",
						Expected: vocab.Tty,
					})
				}
			}

			if !codeSet[strings.ToLower(vocab.Code)] {
				diff = append(diff, Difference{
					Reason:   "Such CODE was not found",
					Expected: vocab.Code,
				})
			}
		}

	}
	return diff
}

type SabConcept struct {
	CodingScheme  string         `json:"codingScheme"`
	VocabConcepts []VocabConcept `json:"vocabConcepts"`
}

type VocabConcept struct {
	Tty  []string `json:"tty"`
	Code string   `json:"code"`
}

type Attributes struct {
	LabValue           Span   `json:"labValue,omitempty"`
	LabValueUnit       Span   `json:"labValueUnit,omitempty"`
	MedFrequencyNumber Span   `json:"medFrequencyNumber,omitempty"`
	MedFrequencyUnit   Span   `json:"medFrequencyUnit,omitempty"`
	MedStrengthNum     Span   `json:"medStrengthNum,omitempty"`
	MedStrengthUnit    Span   `json:"medStrengthUnit,omitempty"`
	MedStatusChange    string `json:"medStatusChange,omitempty"`
	MedDosage          string `json:"medDosage,omitempty"`
	MedRoute           string `json:"medRoute,omitempty"`
	MedForm            string `json:"medForm,omitempty"`
	MedDuration        string `json:"medDuration,omitempty"`
	Polarity           string `json:"polarity,omitempty"`
}

type Span []interface{}

func (orig Span) Equal(comp Span) bool {
	switch {
	case len(orig) == 0:
		return len(orig) == len(comp)
	case len(orig) > len(comp):
		return false
	case len(orig) == 1:
		return strings.EqualFold(orig[0].(string), comp[0].(string))
	}

	textOrig, beginOrig, endOrig := orig[0].(string), orig[1].(float64), orig[2].(float64)
	text, begin, end := comp[0].(string), comp[1].(float64), comp[2].(float64)

	return strings.EqualFold(textOrig, text) && beginOrig == begin && endOrig == end
}

type Testdata struct {
	ReportPath     string
	CtakesResultCh <-chan CtakesFileResult
	Pipeline       Pipeline
}

type Results struct {
	filename string
	data     interface{}
}

type MTSample struct {
	filename string
	body     string
}

func NewTestdata() Testdata {
	t := Testdata{}

	rootPath, err := filepath.Abs("../../")
	if err != nil {
		return t
	}

	t.ReportPath = path.Join(rootPath, "testdata/reports/")

	cfgs, err := types.LoadConfigurations(path.Join(rootPath, "config"))
	if err != nil {
		return t
	}

	params := GetDefaultClinicalParams(rootPath, path.Join(rootPath, "resources", "dictionaries"), cfgs)
	t.Pipeline, err = DefaultClinical(params)
	if err != nil {
		return t
	}

	t.CtakesResultCh = func() <-chan CtakesFileResult {

		sampleCh := make(chan MTSample)
		go func() {
			defer close(sampleCh)

			samplesPath := path.Join(rootPath, "testdata/mtsamples-clean.zip")
			mtReader, err := zip.OpenReader(path.Join(rootPath, "testdata/mtsamples-clean.zip"))
			if err != nil {
				fmt.Printf("Failed OpenReader of %s, %s", samplesPath, err)
				return
			}
			defer mtReader.Close()

			for _, sample := range mtReader.File {
				if filepath.Ext(sample.Name) != ".txt" {
					continue
				}
				reader, err := sample.Open()
				if err != nil {
					fmt.Printf("Error Open of %s, %s", sample.Name, err)
					return
				}
				buf, err := ioutil.ReadAll(reader)
				if err != nil {
					fmt.Printf("Error ReadAll of %s, %s", sample.Name, err)
					return
				}
				err = reader.Close()
				if err != nil {
					fmt.Printf("Error Close of %s, %s", sample.Name, err)
					return
				}
				sampleCh <- MTSample{
					filename: filepath.Base(sample.Name),
					body:     string(buf),
				}
			}
		}()

		ctakesResCh := make(chan CtakesFileResult)
		go func() {
			defer close(ctakesResCh)

			pathSource := path.Join(rootPath, "testdata/ctakes.zip")
			r, err := zip.OpenReader(pathSource)
			if err != nil {
				fmt.Printf("Failed OpenReader of %s, %s", pathSource, err)
				return
			}
			defer r.Close()

			for sample := range sampleCh {
				ctakesRes := CtakesFileResult{
					Filename: sample.filename,
					Text:     sample.body,
				}

				allPipelines := make([][]byte, 0, 3)
				for _, item := range r.File {
					// Parse file
					type ctakesFile struct {
						file     *zip.File
						config   string
						pipeline string
						filename string
					}
					file, err := func(file *zip.File) (ctakesFile, error) {
						var f ctakesFile

						if !strings.EqualFold(filepath.Ext(file.Name), ".json") {
							return f, fmt.Errorf("wrong format, %s", filepath.Ext(f.filename))
						}
						f.filename = strings.TrimSuffix(filepath.Base(file.Name), filepath.Ext(file.Name))
						if f.filename != sample.filename {
							return f, fmt.Errorf("skipping")
						}
						f.file = file
						f.pipeline = filepath.Base(filepath.Dir(file.Name))
						f.config = filepath.Base(filepath.Dir(filepath.Dir(file.Name)))

						return f, nil
					}(item)
					if err != nil {
						continue
					}

					r, _ := file.file.Open()
					buf, _ := ioutil.ReadAll(r)

					if strings.EqualFold(file.config, "hepc") {
						allPipelines = append(allPipelines, buf)
						if len(allPipelines) < 3 {
							continue
						}
					}
					if len(allPipelines) != 0 {
						var subBuf []byte
						for i := range allPipelines {
							if subBuf == nil {
								subBuf = allPipelines[i]
								continue
							}
							subBuf, err = jsonpatch.MergePatch(allPipelines[i], subBuf)
							if err != nil {
								fmt.Printf("Error in MergePatch")
							}
						}
						buf = subBuf
					}

					var res TestResponse
					err = json.Unmarshal(buf, &res)
					if err != nil {
						fmt.Printf("Error in Unmarshal, %s", err)
					}

					ctakesRes.Results = append(ctakesRes.Results, AnnotationResult{
						Config:   file.config,
						Response: res,
					})
				}

				ctakesResCh <- ctakesRes
			}
		}()
		return ctakesResCh
	}()

	return t
}
