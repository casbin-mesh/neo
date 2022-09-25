package utils

import (
	"bufio"
	"encoding/csv"
	"github.com/casbin-mesh/neo/pkg/primitive/value"
	"os"
	"strings"
)

func recordsToTuples(src [][]string) (out []value.Values) {
	for _, stringList := range src {
		elems := make([]value.Value, 0, len(stringList))
		for _, s := range stringList {
			elems = append(elems, value.NewStringValue(strings.TrimSpace(s)))
		}
		out = append(out, elems)
	}
	return
}

func CsvToTuples(path string) ([]value.Values, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return recordsToTuples(data), nil
}

func ReadFile(path string) (*bufio.Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(f), nil
}
