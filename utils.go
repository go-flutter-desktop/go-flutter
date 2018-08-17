package main

import (
	"unicode"
)

// From https://github.com/peterh/liner/blob/8c1271fcf47f341a9e6771872262870e1ad7650c/line.go#L819-L837

func deleteLeadingWord(line []rune) []rune {

	pos := len(line)

	// Remove whitespace to the left
	var buf []rune // Store the deleted chars in a buffer
	for {
		if pos == 0 || !unicode.IsSpace(line[pos-1]) {
			break
		}
		buf = append(buf, line[pos-1])
		line = append(line[:pos-1], line[pos:]...)
		pos--
	}

	// Remove non-whitespace to the left
	for {
		if pos == 0 || unicode.IsSpace(line[pos-1]) {
			break
		}
		buf = append(buf, line[pos-1])
		line = append(line[:pos-1], line[pos:]...)
		pos--
	}

	return line

}
