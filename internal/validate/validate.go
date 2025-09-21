package validate

import (
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// Struct validates a struct using tags.
func Struct(s any) error {
	return v.Struct(s)
}

// Var validates a single value against a tag.
func Var(field any, tag string) error {
	return v.Var(field, tag)
}
