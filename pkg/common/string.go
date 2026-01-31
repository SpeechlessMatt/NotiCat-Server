// Package common some general tools
package common

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-20

import (
	"encoding/json"
	"regexp"
	"strings"
)

func ParseJSONString(data string) (any, error) {
	if data == "" {
		return make(map[string]any), nil
	}

	var result any
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return data, err
	}
	return result, nil
}

func NormalizeJSON(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}

	b, _ := json.Marshal(m)
	return string(b)
}

func ShortenTitle(s string) string {
	cleanText := strings.NewReplacer("\n", "", "\r", "").Replace(s)
	runes := []rune(cleanText)
	limit := min(20, len(runes))
	return string(runes[:limit])
}

type StringFilter struct {
	Pattern    string
	IsRegexp   bool
	IgnoreCase bool
	re         *regexp.Regexp
}

func NewFilter(pattern string, isRegexp bool, ignoreCase bool) (*StringFilter, error) {
	filter := &StringFilter{
		Pattern:    pattern,
		IsRegexp:   isRegexp,
		IgnoreCase: ignoreCase,
	}

	if isRegexp {
		finalPattern := pattern
		if ignoreCase {
			finalPattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(finalPattern)
		if err != nil {
			return nil, err
		}
		filter.re = re
	}

	return filter, nil
}

func (f *StringFilter) Match(input string) bool {
	if f.IsRegexp {
		return f.re.MatchString(input)
	}

	// 普通字符串逻辑
	if f.IgnoreCase {
		return strings.Contains(strings.ToLower(input), strings.ToLower(f.Pattern))
	}
	return strings.Contains(input, f.Pattern)
}
