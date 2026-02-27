package index

import (
	"regexp"
	"sort"
)

func SortResults(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Note.UpdatedAt.After(results[j].Note.UpdatedAt)
		}
		return results[i].Score > results[j].Score
	})
}

func CompileRegex(query string) (*regexp.Regexp, error) {
	return regexp.Compile(query)
}
