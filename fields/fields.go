package fields

import (
	"strings"
)

type Checker interface {
	FieldProvided(string) bool
}

type FieldChecker struct {
	Args map[string]interface{}
}

func (n *FieldChecker) FieldProvided(field string) bool {
	path := strings.Split(field, ".")
	return step(n.Args, path)
}

func untitle(s string) string {
	return strings.Join([]string{
		strings.ToLower(s[:1]),
		s[1:],
	}, "")
}

func step(args map[string]interface{}, path []string) bool {
	if val, ok := args[untitle(path[0])]; ok {
		if len(path) == 1 {
			return true
		}
		if v, ok := val.(map[string]interface{}); ok {
			return step(v, path[1:])
		}

		return true
	}
	return false
}
