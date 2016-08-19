package main

import (
	"html/template"
	"strings"
)

func split(s string, d string) []string {
	arr := strings.Split(s, d)
	return arr
}

func in(haystack interface{}, needle string) bool {
	if s, ok := haystack.(string); ok {
		return strings.Contains(s, needle)
	}

	return false
}

var helperFuncs = template.FuncMap{
	"split": split,
	"in":    in,
}
