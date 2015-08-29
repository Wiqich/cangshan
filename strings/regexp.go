package stringutil

import "regexp"

func MatchRegexpMap(exp *regexp.Regexp, text string) map[string]string {
	submatch := exp.FindStringSubmatch(text)
	if submatch == nil {
		return nil
	}
	match := make(map[string]string)
	for i, name := range exp.SubexpNames() {
		if name != "" && submatch[i] != "" {
			match[name] = submatch[i]
		}
	}
	return match
}

func MatchAllRegexpMap(exp *regexp.Regexp, text string, n int) []map[string]string {
	submatches := exp.FindAllStringSubmatch(text, n)
	if submatches == nil {
		return nil
	}
	matches := make([]map[string]string, len(submatches))
	for i, submatch := range submatches {
		matches[i] = make(map[string]string)
		for j, name := range exp.SubexpNames() {
			if name != "" && submatch[j] != "" {
				matches[i][name] = submatch[j]
			}
		}
	}
	return matches
}
