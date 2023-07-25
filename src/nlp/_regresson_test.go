package nlp

import (
	"text2phenotype.com/fdl/types"
	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

type RegressionEnvConfig struct {
	ResourcesDirPath   string `envconfig:"FDL_RESOURCES_PATH" required:"true"` // Should point to a directory with FDL resources
	TextSamplesDirPath string `envconfig:"TEXT_SAMPLES_PATH" required:"true"`  // Should point to a directory containing sample texts in .txt format
	FixturesDirPath    string `envconfig:"FIXTURES_PATH" required:"true"`      // Should point to a directory containing fixtures in .json format with []SpanFixture in each file
}

type SpanFixture struct {
	Begin     int32  `json:"b"`
	End       int32  `json:"e"`
	FirstWord string `json:"fw"`
	LastWord  string `json:"lw"`
}

// This test requires text samples to run, so it's primarily intended to be run in local dev environments.
// In order to run this test several environment variables listed in RegressionEnvConfig structure should be set.
//
// text2phenotype-samples were used to port SD from CTAKES and check new SD implementation
func Test_EnsureSameResults(t *testing.T) {
	var config RegressionEnvConfig
	err := envconfig.Process("", &config)
	if err != nil {
		t.Fatal(err)
	}
	resPath := config.ResourcesDirPath
	detectorResources := path.Join(resPath, "sentdetector")
	detector, err := NewSentenceDetector(detectorResources)
	if err != nil {
		t.Fatal(err)
	}
	samplesPath := resolvePath(t, config.TextSamplesDirPath)
	fixturesPath := resolvePath(t, config.FixturesDirPath)

	err = filepath.WalkDir(samplesPath, func(path string, d fs.DirEntry, err1 error) error {
		if err1 != nil {
			t.Log(err1)
			return err1
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".txt") {
			return nil
		}
		t.Log("Starting", name)
		text, err := loadText(filepath.Join(samplesPath, name))
		if err != nil {
			t.Fatal(err, name)
		}
		newResults := runSD(detector, text)
		fixtureName := strings.TrimSuffix(name, "txt") + "json"
		compareResultsWithFixture(t, name, loadFixture(t, fixtureName, fixturesPath), newResults)

		t.Log("Finished", name)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func resolvePath(t *testing.T, basePath string) string {
	dirInfo, err := os.Lstat(basePath)
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode()&os.ModeSymlink == 0 {
		return basePath
	}
	readlink, err := os.Readlink(basePath)
	if err != nil {
		t.Fatal(err)
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

func runSD(sd func(in <-chan string) <-chan types.Sentence, text string) []types.Sentence {
	in := make(chan string)
	out := sd(in)
	in <- text
	close(in)
	results := make([]types.Sentence, 0, 300)
	for sentence := range out {
		results = append(results, sentence)
	}
	return results
}

func compareResultsWithFixture(t *testing.T, name string, expected []SpanFixture, actual []types.Sentence) {
	require.Equal(t, len(expected), len(actual), "Failed %s", name)

	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		act := actual[i]
		words := strings.Split(*act.Text, " ")
		require.Equal(t, exp.FirstWord, words[0], "Failed %s", name)
		require.Equal(t, exp.LastWord, words[len(words)-1], "Failed %s", name)
		require.Equal(t, exp.Begin, act.Begin, "Failed %s", name)
		require.Equal(t, exp.End, act.End, "Failed %s", name)
	}
}

func saveFixture(t *testing.T, name, dir string, sentences []types.Sentence) {
	spans := make([]SpanFixture, len(sentences))
	for i, sentence := range sentences {
		words := strings.Split(*sentence.Text, " ")
		spans[i] = SpanFixture{
			Begin:     sentence.Begin,
			End:       sentence.End,
			FirstWord: words[0],
			LastWord:  words[len(words)-1],
		}
	}
	data, err := json.Marshal(spans)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, name), data, 0666); err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, name, dir string) []SpanFixture {
	var spans []SpanFixture
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(data, &spans)
	if err != nil {
		t.Fatal(err)
	}
	return spans
}
