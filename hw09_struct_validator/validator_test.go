package hw09structvalidator

import (
	"errors"
	"testing"
)

type UserRole string

type (
	User struct {
		ID     string   `validate:"len:36"`
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int `validate:"in:200,404,500"`
	}

	BadTags struct {
		Age int `validate:"min"`
	}

	BadRegexp struct {
		Email string `validate:"regexp:("`
	}
)

func TestValidate_OK(t *testing.T) {
	t.Parallel()

	u := User{
		ID:     "12345678-1234-1234-1234-123456789012",
		Age:    30,
		Email:  "john@doe.com",
		Role:   "admin",
		Phones: []string{"12345678901", "09876543210"},
	}

	requireNoErr(t, Validate(u))
}

func TestValidate_MultipleErrors(t *testing.T) {
	t.Parallel()

	u := User{
		ID:     "short",
		Age:    10,
		Email:  "bad-email",
		Role:   "user",
		Phones: []string{"123", "12345678901", "000"},
	}

	err := Validate(u)
	verrs := requireValidationErrors(t, err)

	if len(verrs) != 6 {
		t.Fatalf("expected 6 validation errors, got %d: %v", len(verrs), verrs)
	}

	assertFieldCounts(t, verrs, map[string]int{
		"ID":     1,
		"Age":    1,
		"Email":  1,
		"Role":   1,
		"Phones": 2,
	})

	assertHasKinds(t, verrs, []error{ErrLen, ErrMin, ErrRegexp, ErrIn})
}

func TestValidate_SingleErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("app version invalid len", func(t *testing.T) {
		t.Parallel()

		err := Validate(App{Version: "1.0"})
		verrs := requireValidationErrors(t, err)

		if len(verrs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(verrs))
		}
		if verrs[0].Field != "Version" || !errors.Is(verrs[0].Err, ErrLen) {
			t.Fatalf("unexpected error: %v", verrs[0])
		}
	})

	t.Run("response code not in", func(t *testing.T) {
		t.Parallel()

		err := Validate(Response{Code: 201})
		verrs := requireValidationErrors(t, err)

		if len(verrs) != 1 {
			t.Fatalf("expected 1 error, got %d", len(verrs))
		}
		if verrs[0].Field != "Code" || !errors.Is(verrs[0].Err, ErrIn) {
			t.Fatalf("unexpected error: %v", verrs[0])
		}
	})
}

func TestValidate_IgnoresUntaggedFields(t *testing.T) {
	t.Parallel()
	requireNoErr(t, Validate(Token{}))
}

func TestValidate_ProgramErrors(t *testing.T) {
	t.Parallel()

	t.Run("non-struct input", func(t *testing.T) {
		t.Parallel()

		if !errors.Is(Validate("not a struct"), ErrNotStruct) {
			t.Fatalf("expected ErrNotStruct")
		}
	})

	t.Run("bad tag format", func(t *testing.T) {
		t.Parallel()

		if !errors.Is(Validate(BadTags{Age: 10}), ErrInvalidTag) {
			t.Fatalf("expected ErrInvalidTag")
		}
	})

	t.Run("bad regexp", func(t *testing.T) {
		t.Parallel()

		if !errors.Is(Validate(BadRegexp{Email: "x"}), ErrInvalidRegexp) {
			t.Fatalf("expected ErrInvalidRegexp")
		}
	})
}

func requireNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func requireValidationErrors(t *testing.T, err error) ValidationErrors {
	t.Helper()

	var verrs ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}
	return verrs
}

func assertFieldCounts(t *testing.T, verrs ValidationErrors, want map[string]int) {
	t.Helper()

	got := map[string]int{}
	for _, ve := range verrs {
		got[ve.Field]++
	}

	for f, c := range want {
		if got[f] != c {
			t.Fatalf("field %s: expected %d errs, got %d (all=%v)", f, c, got[f], verrs)
		}
	}
}

func assertHasKinds(t *testing.T, verrs ValidationErrors, wantKinds []error) {
	t.Helper()

	present := map[error]bool{}
	for _, ve := range verrs {
		for _, k := range wantKinds {
			if errors.Is(ve.Err, k) {
				present[k] = true
			}
		}
	}

	for _, k := range wantKinds {
		if !present[k] {
			t.Fatalf("expected error kind %v to be present; got=%v", k, verrs)
		}
	}
}
