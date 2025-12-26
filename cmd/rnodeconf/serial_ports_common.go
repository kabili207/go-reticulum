package main

import (
	"path/filepath"
	"sort"
)

func globSerialPatterns(patterns []string) []string {
	seen := map[string]struct{}{}
	var ports []string
	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		for _, m := range matches {
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			ports = append(ports, m)
		}
	}
	sort.Strings(ports)
	return ports
}

