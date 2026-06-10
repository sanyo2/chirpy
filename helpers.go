package main

import (
	"strings"
)

func replaceBadWords(msg string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	var allWords []string

	for _, word := range strings.Split(msg, " ") {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			allWords = append(allWords, "****")
			continue
		}
		allWords = append(allWords, word)
	}

	return strings.Join(allWords, " ")
}
