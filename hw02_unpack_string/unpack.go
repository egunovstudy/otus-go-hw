package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	var builder strings.Builder
	var currentState = START
	var lastRune rune

	for _, char := range input {
		newState := defineState(char)
		if err := checkTransitionAllowed(currentState, newState); err != nil {
			return "invalid string", err
		}

		if newState == CHAR {
			if lastRune != 0 {
				builder.WriteRune(lastRune)
			}
			lastRune = char
		} else if newState == DIGIT {
			builder.WriteString(strings.Repeat(string(lastRune), int(char-'0')))
			lastRune = 0
		}
		currentState = newState
	}
	if lastRune != 0 {
		builder.WriteRune(lastRune)
	}

	return builder.String(), nil
}

func checkTransitionAllowed(current, inputState InputState) error {
	if !isTransitionAllowed(current, inputState) {
		return ErrInvalidString
	}
	return nil
}

func isTransitionAllowed(from InputState, to InputState) bool {
	if from == START {
		return to == CHAR
	}
	if from == CHAR {
		return to == CHAR || to == DIGIT
	}
	if from == DIGIT {
		return to == CHAR
	}
	return false
}

func defineState(char int32) InputState {
	var inpSt InputState
	if unicode.IsNumber(char) {
		inpSt = DIGIT
	} else if unicode.IsLetter(char) || unicode.IsGraphic(char) {
		inpSt = CHAR
	}
	return inpSt
}

type InputState string

const (
	START InputState = "START"
	CHAR  InputState = "CHAR"
	DIGIT InputState = "DIGIT"
)
