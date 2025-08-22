package hjson

import (
	"encoding/json"
	"errors"
	"math"
	"strconv"
)

type parseNumber struct {
	data []byte
	at   int  // The index of the current character
	ch   byte // The current character
}

func (p *parseNumber) next() bool {
	// get the next character.
	len := len(p.data)
	if p.at < len {
		p.ch = p.data[p.at]
		p.at++
		return true
	}
	if p.at == len {
		p.at++
		p.ch = 0
	}
	return false
}

func (p *parseNumber) peek(offs int) byte {
	pos := p.at + offs
	if pos >= 0 && pos < len(p.data) {
		return p.data[pos]
	}
	return 0
}

func startsWithNumber(text []byte) bool {
	if _, err := tryParseNumber(text, true, false); err == nil {
		return true
	}
	return false
}

func tryParseNumber(text []byte, stopAtNext, useJSONNumber bool) (interface{}, error) {
	// Parse a number value.

	p := parseNumber{
		data: text,
		at:   0,
		ch:   ' ',
	}
	leadingZeros := 0
	testLeading := true
	p.next()
	if p.ch == '-' {
		p.next()
	}
	for p.ch >= '0' && p.ch <= '9' {
		if testLeading {
			if p.ch == '0' {
				leadingZeros++
			} else {
				testLeading = false
			}
		}
		p.next()
	}
	if testLeading {
		leadingZeros--
	} // single 0 is allowed
	if p.ch == '.' {
		for p.next() && p.ch >= '0' && p.ch <= '9' {
		}
	}
	if p.ch == 'e' || p.ch == 'E' {
		p.next()
		if p.ch == '-' || p.ch == '+' {
			p.next()
		}
		for p.ch >= '0' && p.ch <= '9' {
			p.next()
		}
	}

	end := p.at

	// skip white/to (newline)
	for p.ch > 0 && p.ch <= ' ' {
		p.next()
	}

	if stopAtNext {
		// end scan if we find a punctuator character like ,}] or a comment
		if p.ch == ',' || p.ch == '}' || p.ch == ']' ||
			p.ch == '#' || p.ch == '/' && (p.peek(0) == '/' || p.peek(0) == '*') {
			p.ch = 0
		}
	}

	if p.ch > 0 || leadingZeros != 0 {
		return 0, errors.New("Invalid number")
	}
	if useJSONNumber {
		return json.Number(string(p.data[0 : end-1])), nil
	}
	number, err := strconv.ParseFloat(string(p.data[0:end-1]), 64)
	if err != nil {
		return 0, err
	}
	if math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, errors.New("Invalid number")
	}
	return number, nil
}
