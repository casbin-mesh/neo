package utils

import (
	"encoding/csv"
	"github.com/casbin-mesh/neo/pkg/primitive/btuple"
	"os"
)

func recordsToTuples(src [][]string) (out []btuple.Modifier) {
	for _, strings := range src {
		elems := make([]btuple.Elem, 0, len(strings))
		for _, s := range strings {
			elems = append(elems, []byte(s))
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
