// Package hw09structvalidator provides struct field validation by `validate` tags.
package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError describes a single validation error for a field.
type ValidationError struct {
	Field string
	Err   error
}

// ValidationErrors is a list of ValidationError.
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	parts := make([]string, 0, len(v))
	for _, ve := range v {
		parts = append(parts, fmt.Sprintf("%s: %v", ve.Field, ve.Err))
	}

	return strings.Join(parts, "; ")
}

var (
	// ErrNotStruct is returned when Validate receives a non-struct value.
	ErrNotStruct = errors.New("expected struct")

	// ErrInvalidTag is returned when `validate` tag has invalid format.
	ErrInvalidTag = errors.New("invalid validate tag")

	// ErrUnknownValidator is returned when an unknown validator is used in a tag.
	ErrUnknownValidator = errors.New("unknown validator")

	// ErrInvalidValue is returned when validator value cannot be parsed.
	ErrInvalidValue = errors.New("invalid validator value")

	// ErrInvalidRegexp is returned when regexp validator contains invalid regexp.
	ErrInvalidRegexp = errors.New("invalid regexp")

	// ErrUnsupportedType is returned when validator is applied to unsupported field type.
	ErrUnsupportedType = errors.New("unsupported field type")

	// ErrLen indicates `len` validation failure.
	ErrLen = errors.New("validation failed: len")

	// ErrRegexp indicates `regexp` validation failure.
	ErrRegexp = errors.New("validation failed: regexp")

	// ErrIn indicates `in` validation failure.
	ErrIn = errors.New("validation failed: in")

	// ErrMin indicates `min` validation failure.
	ErrMin = errors.New("validation failed: min")

	// ErrMax indicates `max` validation failure.
	ErrMax = errors.New("validation failed: max")
)

// Validate validates exported struct fields according to `validate` tags.
func Validate(v interface{}) error {
	if v == nil {
		return ErrNotStruct
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ErrNotStruct
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	rt := rv.Type()
	var all ValidationErrors

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}

		tag, ok := sf.Tag.Lookup("validate")
		if !ok || strings.TrimSpace(tag) == "" {
			continue
		}

		valErrs, err := validateField(rv.Field(i), tag)
		if err != nil {
			return fmt.Errorf("%w: field %s", err, sf.Name)
		}
		for _, ve := range valErrs {
			all = append(all, ValidationError{Field: sf.Name, Err: ve})
		}
	}

	if len(all) > 0 {
		return all
	}

	return nil
}

type rule struct {
	name  string
	value string
}

func validateField(v reflect.Value, tag string) ([]error, error) {
	rules, err := parseRules(tag)
	if err != nil {
		return nil, err
	}

	if v.Kind() == reflect.Slice {
		return validateSlice(v, rules)
	}

	return validateScalar(v, rules)
}

func parseRules(tag string) ([]rule, error) {
	raw := strings.Split(tag, "|")
	parsed := make([]rule, 0, len(raw))

	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			return nil, ErrInvalidTag
		}

		key, val, ok := strings.Cut(r, ":")
		if !ok {
			return nil, ErrInvalidTag
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" || val == "" {
			return nil, ErrInvalidTag
		}

		parsed = append(parsed, rule{name: key, value: val})
	}

	return parsed, nil
}

func validateSlice(v reflect.Value, rules []rule) ([]error, error) {
	if v.Kind() != reflect.Slice {
		return nil, ErrUnsupportedType
	}

	ek := v.Type().Elem().Kind()
	if !isStringKind(ek) && !isIntKind(ek) {
		return nil, ErrUnsupportedType
	}

	var errs []error
	for i := 0; i < v.Len(); i++ {
		valErrs, err := validateScalar(v.Index(i), rules)
		if err != nil {
			return nil, err
		}
		errs = append(errs, valErrs...)
	}

	return errs, nil
}

func validateScalar(v reflect.Value, rules []rule) ([]error, error) {
	k := v.Kind()

	if !isStringKind(k) && !isIntKind(k) {
		return nil, ErrUnsupportedType
	}

	var errs []error

	for _, r := range rules {
		if isStringKind(k) {
			verr, err := validateStringRule(v.String(), r)
			if err != nil {
				return nil, err
			}
			if verr != nil {
				errs = append(errs, verr)
			}
			continue
		}

		verr, err := validateIntRule(v.Int(), r)
		if err != nil {
			return nil, err
		}
		if verr != nil {
			errs = append(errs, verr)
		}
	}

	return errs, nil
}

func isIntKind(k reflect.Kind) bool {
	return k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64
}

func isStringKind(k reflect.Kind) bool {
	return k == reflect.String
}

func validateStringRule(s string, r rule) (validationErr error, programErr error) {
	switch r.name {
	case "len":
		n, err := strconv.Atoi(r.value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidValue, err)
		}
		if len(s) != n {
			return fmt.Errorf("%w: expected %d, got %d", ErrLen, n, len(s)), nil
		}
		return nil, nil

	case "regexp":
		re, err := regexp.Compile(r.value)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidRegexp, err)
		}
		if !re.MatchString(s) {
			return fmt.Errorf("%w: %s", ErrRegexp, r.value), nil
		}
		return nil, nil

	case "in":
		allowed := strings.Split(r.value, ",")
		for i := range allowed {
			allowed[i] = strings.TrimSpace(allowed[i])
		}
		for _, a := range allowed {
			if s == a {
				return nil, nil
			}
		}
		return fmt.Errorf("%w: %s", ErrIn, r.value), nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownValidator, r.name)
	}
}

func validateIntRule(n int64, r rule) (validationErr error, programErr error) {
	switch r.name {
	case "min":
		minV, err := strconv.ParseInt(r.value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidValue, err)
		}
		if n < minV {
			return fmt.Errorf("%w: expected >= %d, got %d", ErrMin, minV, n), nil
		}
		return nil, nil

	case "max":
		maxV, err := strconv.ParseInt(r.value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidValue, err)
		}
		if n > maxV {
			return fmt.Errorf("%w: expected <= %d, got %d", ErrMax, maxV, n), nil
		}
		return nil, nil

	case "in":
		parts := strings.Split(r.value, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			v, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: %w", ErrInvalidValue, err)
			}
			if n == v {
				return nil, nil
			}
		}
		return fmt.Errorf("%w: %s", ErrIn, r.value), nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownValidator, r.name)
	}
}
