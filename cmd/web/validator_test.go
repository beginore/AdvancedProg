// validator_test.go
package main

import (
	"forum/internal/validator"
	"testing"
)

func TestValidator_NotBlank(t *testing.T) {
	v := validator.Validator{}
	v.CheckField(validator.NotBlank(""), "field", "Cannot be blank")

	if v.Valid() {
		t.Error("Expected validation to fail for blank field, but it passed")
	}

	v = validator.Validator{}
	v.CheckField(validator.NotBlank("not blank"), "field", "Cannot be blank")
	if !v.Valid() {
		t.Error("Expected validation to pass for non-blank field, but it failed")
	}
}
