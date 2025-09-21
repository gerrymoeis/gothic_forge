package validate

import "testing"

type sample struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18,lte=80"`
}

func TestStruct_Valid(t *testing.T) {
	s := sample{Email: "user@example.com", Age: 30}
	if err := Struct(s); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestStruct_Invalid(t *testing.T) {
	s := sample{Email: "bad", Age: 10}
	if err := Struct(s); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestVar(t *testing.T) {
	if err := Var("user@example.com", "required,email"); err != nil {
		t.Fatalf("expected valid email, got %v", err)
	}
	if err := Var("oops", "email"); err == nil {
		t.Fatalf("expected invalid email error")
	}
}
