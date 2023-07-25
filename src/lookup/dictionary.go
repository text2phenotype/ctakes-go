package lookup

import (
	"bufio"
	"text2phenotype.com/fdl/logger"
	"text2phenotype.com/fdl/types"
	"text2phenotype.com/fdl/utils"
	"crypto/sha256"
	"encoding/hex"
	"github.com/rs/zerolog"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Dictionary func(words []*string) MapListIterator

func CreateDictionary(configName string, path string, scheme []string) (Dictionary, error) {
	fdlLogger := logger.NewLogger("Dictionary loader").With().
		Str("config_name", configName).
		Str("path", path).Logger()
	errLogger := fdlLogger.With().Caller().Logger()
	fdlLogger.Info().Msg("Started loading")

	idxCachePath, err := getDstFilepath(path, scheme, &errLogger)
	if err != nil {
		errLogger.Err(err).Msg("Could not create index cache path")
		return nil, err
	}
	fdlLogger = fdlLogger.With().Str("index_cache_path", idxCachePath).Logger()
	errLogger = errLogger.With().Str("index_cache_path", idxCachePath).Logger()

	indexExists := func() bool {
		_, err := os.Stat(idxCachePath)
		return err == nil
	}()

	rwMap := make(RareWordTermMap)

	if !indexExists {
		fdlLogger.Info().Msg("Building new index")
		var schemaMap = make(map[string]byte)
		for i, columnName := range scheme {
			schemaMap[columnName] = byte(i)
		}

		cuiIdx := schemaMap[types.CUI]
		strIdx := schemaMap[types.STR]

		tokenizr, err := NewTermTokenizer()
		if err != nil {
			errLogger.Err(err).Msg("Could not create term tokenizer")
			return nil, err
		}

		var rare_words []*RareWordTerm

		getHash := func(columns []string) uint64 {
			cui := columns[cuiIdx]
			term := columns[strIdx]
			return utils.HashString(cui + "_" + term)
		}

		reader, err := utils.NewBSVReader(path, getHash)
		if err != nil {
			errLogger.Err(err).Msg("Could not create BSV reader")
			return nil, err
		}

		for columns := range reader {
			cui := columns[cuiIdx]
			term := columns[strIdx]

			tokenizedTerm := tokenizr(term)

			if len(tokenizedTerm) == 0 {
				continue
			}

			tokenPointers := make([]*string, len(tokenizedTerm))
			for i, token := range tokenizedTerm {
				tokenPointers[i] = utils.GlobalStringStore().GetPointer(token)
			}

			intst := RareWordTerm{
				Tokens:        tokenPointers,
				CUI:           utils.GlobalStringStore().GetPointer(cui),
				TextLength:    uint32(len(term)),
				RareWordIndex: 0,
			}

			rare_words = append(rare_words, &intst)
		}

		rwMap = createRareWordMap(rare_words)
		serialized, err := rwMap.MarshalJSON()
		if err != nil {
			errLogger.Err(err).Msg("Got error while marshalling rare words map")
			return nil, err
		}

		go func(data []byte) {
			err := os.MkdirAll(filepath.Dir(idxCachePath), 0700)
			if err != nil {
				errLogger.Err(err).Msg("Could not create dictionary for index cache")
				return
			}
			f, err := os.Create(idxCachePath)
			if err != nil {
				errLogger.Err(err).Msg("Could not create file for index cache")
			}
			defer func(f *os.File) {
				err := f.Close()
				if err != nil {
					errLogger.Err(err).Msg("Caught error while closing cache file")
				}
			}(f)
			w := bufio.NewWriter(f)
			_, err = w.Write(data)
			if err != nil {
				errLogger.Err(err).Msg("Could not write serialized rare words map")
			}
		}(serialized)

	} else {
		fdlLogger.Info().Msg("Loading index from cache")
		indCache, err := ioutil.ReadFile(idxCachePath)
		if err != nil {
			return nil, err
		}

		err = rwMap.UnmarshalJSON(indCache)
		if err != nil {
			return nil, err
		}
	}

	fdlLogger.Info().Msgf("%d terms were loaded", len(rwMap))
	return func(words []*string) MapListIterator {
		return CreateMapListIterator(rwMap, words)
	}, nil
}

func getDstFilepath(dictPath string, scheme []string, errLogger *zerolog.Logger) (string, error) {
	hash, err := func() (string, error) {
		f, err := os.Open(dictPath)
		if err != nil {
			return "", err
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		hasher := sha256.New()
		if _, err := io.Copy(hasher, f); err != nil {
			return "", err
		}
		result := utils.HashString(strings.Join(scheme, "") + hex.EncodeToString(hasher.Sum(nil)))
		return strconv.FormatUint(result, 10), nil
	}()

	if err != nil {
		errLogger.Err(err).Msg("Could not read dictionary file")
		return "", err
	}
	resourcePath := filepath.Dir(filepath.Dir(filepath.Dir(dictPath)))
	idxDictDir := filepath.Base(filepath.Dir(dictPath))

	idxName := strings.TrimSuffix(filepath.Base(dictPath), filepath.Ext(dictPath))
	filename := strings.Join([]string{idxName, hash, ".json"}, "")

	return filepath.Join(resourcePath, "tmp_index", idxDictDir, filename), nil
}
