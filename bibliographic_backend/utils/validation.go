package utils

import (
	"errors"
	"regexp"
	"strings"
)

const (
	minSearchSymbols = 5
	minFormatSymbols = 15
	minFormatWords   = 4
	maxSearchSymbols = 100
	maxFormatSymbols = 15000
	maxFormatLinks   = 100
)

var (
	ErrDigitsOnly     = errors.New("input is only digits")
	ErrInputTooShort  = errors.New("input is too short")
	ErrNotEnoughWords = errors.New("not enough words")
	ErrInputTooLong   = errors.New("input is too long")
)

func FormatLinksIsValid(input []string) error {
	if len(input) > maxFormatLinks {
		return ErrInputTooLong
	}

	return nil
}

func SearchInputIsValid(input string) error {
	if len(input) < minSearchSymbols {
		return ErrInputTooShort
	}

	if len(input) > maxSearchSymbols {
		return ErrInputTooLong
	}

	if isDigitsOnly(input) {
		return ErrDigitsOnly
	}

	return nil
}

func FormatInputIsValid(input string) error {
	if len(input) < minFormatSymbols {
		return ErrInputTooShort
	}

	if len(input) > maxFormatSymbols {
		return ErrInputTooLong
	}

	if countWords(input) < minFormatWords {
		return ErrNotEnoughWords
	}

	if isDigitsOnly(input) {
		return ErrDigitsOnly
	}

	return nil
}

func countWords(s string) int {
	words := strings.Fields(s)
	return len(words)
}

func isDigitsOnly(str string) bool {
	match, _ := regexp.MatchString(`^\d+$`, str)
	return match
}
