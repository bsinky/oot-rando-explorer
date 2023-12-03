package randoseed

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func TestValidation(t *testing.T) {
	t.Parallel()

	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		t.Fatal("Validator type assertion failed")
	}

	RegisterValidation(v)

	seed := &Seed{}
	var validationErrors validator.ValidationErrors
	if err := binding.Validator.ValidateStruct(seed); errors.As(err, &validationErrors) {
		if len(validationErrors) == 0 {
			t.Fatal("Should have gotten at least one validation error")
		}
	} else if err != nil {
		t.Fatalf("Error was not expected type %s", err)
	} else if err == nil {
		t.Fatalf("Should have gotten a non-nil error")
	}
}
