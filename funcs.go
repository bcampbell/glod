package main

// this file holds functions for use in templates

import (
	"fmt"
	"html/template"
	"strings"
	"time"
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

// helper func - since we keep all the dates as strings we need to parse
// them more than we should...
func parseDate(d string) (time.Time, error) {
	if d == "" {
		return time.Time{}, fmt.Errorf("blank date/time")
	}
	dateLayouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, fmt := range dateLayouts {
		t, err := time.Parse(fmt, d)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognised date/time: '%s'", d)
}

func dateFormat(fmt string, dt string) (string, error) {
	t, err := parseDate(dt)
	if err != nil {
		return "", err
	}

	return t.Format(fmt), nil
}

var helperFuncs = template.FuncMap{
	"split":      split,
	"in":         in,
	"dateFormat": dateFormat,
}
