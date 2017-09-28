package util

import "strings"

//GetNormalizedVariable transforms an input name into the spec required environment variable
func GetNormalizedVariable(inputName string) string {
	// Remove all non-alphabetic runes, except dash and underscore
	// Upper-case all lower-case alphabetic runes
	// Dash runes are transformed into underscore
	normalizer := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z' || r == '_':
			return r
		case r >= 'a' && r <= 'z':
			return 'A' + (r - 'a')
		case r == '-':
			return '_'
		}
		return -1
	}

	result := strings.Map(normalizer, inputName)

	return result
}

//IsReserved checks if the given string is one of the reserved variable names
func IsReserved(name string, allocated []string) bool {
	reserved := name == "OUTPUT_DIR"

	if allocated != nil {
		for _, s := range allocated {
			if GetNormalizedVariable(s) == strings.ToUpper(name) {
				reserved = true
			}
		}
	}
	return reserved
}

//IsInUse checks if the given string is currently being used by another variable
// Checks if the normalized name is already in use, and if so, adds the path
// so it may be printed later
func IsInUse(name string, path string, vars map[string][]string) bool {
	normName := GetNormalizedVariable(name)

	// normalized name is found in the map.
	if paths, exists := vars[normName]; exists {
		vars[normName] = append(paths, path)
		return true
	}

	// Not found (yet) so add to map
	vars[normName] = []string{path}
	return false
}

//RemoveString returns a new slice of strings without the given string from a slice of strings.
//If the slice does not contain the string the original slice is returned
func RemoveString(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

//ContainsString checks if a string exists in a given slice of strings
func ContainsString(s []string, r string) bool {
	for _, v := range s {
		if v == r {
			return true
		}
	}
	return false
}
