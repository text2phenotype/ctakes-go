package utils

import (
	"bufio"
	"text2phenotype.com/fdl/logger"
	"io"
	"os"
	"path"
	"strings"
)

type GetHashFunc func(columns []string) uint64
type CreateInstanceFunc func(columns []string) error

func NewBSVReader(bsvPath string, getHash GetHashFunc) (<-chan []string, error) {
	_, fileName := path.Split(bsvPath)
	fdlLogger := logger.NewLogger("BSVReader (" + fileName + ")")

	f, err := os.Open(bsvPath)
	if err != nil {
		return nil, err
	}

	out := make(chan []string)

	go func() {
		defer f.Close()
		defer close(out)

		r := bufio.NewReader(f)

		// to remove duplicates
		var hashes = make(map[uint64]bool)

		for {
			line, err := r.ReadString('\n')
			if len(line) == 0 {
				if err == io.EOF {
					break
				} else if err != nil {
					fdlLogger.Error().Err(err)
					return
				}
			}

			if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
			line = strings.ToLower(strings.Trim(line, "\n"))
			columns := strings.Split(line, "|")

			hash := getHash(columns)

			_, ok := hashes[hash]
			if !ok {
				hashes[hash] = true

				out <- columns
			}
		}
	}()

	return out, nil
}
