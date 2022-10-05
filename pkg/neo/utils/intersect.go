package utils

// SortedIntersect Sorted has complexity: O(n * log(n)), a needs to be sorted
func SortedIntersect(a []string, b []string) []string {
	set := make([]string, 0)

	var small, large []string

	if len(a) > len(b) {
		small = b
		large = a
	} else {
		small = a
		large = b
	}

	lookup := map[string]struct{}{}
	// build
	for _, s := range small {
		lookup[s] = struct{}{}
	}
	// probe
	for _, s := range large {
		if _, ok := lookup[s]; ok {
			set = append(set, s)
		}
	}

	return set
}
