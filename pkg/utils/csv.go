package utils

import (
	"encoding/csv"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"os"
	"strings"
)

func recordsToTuples(src [][]string) (out []btuple.Modifier) {
	for _, stringList := range src {
		elems := make([]btuple.Elem, 0, len(stringList))
		for _, s := range stringList {
			elems = append(elems, []byte(strings.TrimSpace(s)))
		}
		out = append(out, btuple.NewModifier(elems))
	}
	return
}

func CsvToTuples(path string) ([]btuple.Modifier, error) {
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
