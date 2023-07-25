package utils

import (
	"bufio"
	"github.com/twmb/murmur3"
	"os"
	"strings"
)

func HashString(s string) uint64 {
	hash := murmur3.New64()
	_, err := hash.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	return hash.Sum64()
}

func HashBytes(bytes ...[]byte) uint64 {
	hash := murmur3.New64()
	for _, b := range bytes {
		_, err := hash.Write(b)
		if err != nil {
			panic(err)
		}
	}
	return hash.Sum64()
}

func HashStrings(ss []string) []uint64 {
	hash := murmur3.New64()

	hashes := make([]uint64, len(ss))
	for i, s := range ss {
		hash.Reset()
		_, err := hash.Write([]byte(s))
		if err != nil {
			panic(err)
		}
		hashes[i] = hash.Sum64()
	}

	return hashes
}

func AbsInt(n int) int {
	if n >= 0 {
		return n
	}

	return -n
}

func ReadMap(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	result := make(map[string]string)
	for scanner.Scan() {
		p := strings.Split(scanner.Text(), "|")
		result[p[0]] = p[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadSet(filePath string) (map[string]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	result := make(map[string]bool)
	for scanner.Scan() {
		result[scanner.Text()] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func ReadList(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var result []string
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
