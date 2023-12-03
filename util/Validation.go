package util

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	ValUnsupportedSeedVersion = "Unknown or unsupported Seed Version"
)

type SimpleValidation struct {
	Message string
}

func (u *SimpleValidation) Error() string {
	return u.Message
}

func (u *SimpleValidation) FieldName() string {
	return ""
}

func friendlyError(err validator.FieldError) string {
	switch err.Tag() {
	case "validVersion":
		return ValUnsupportedSeedVersion
	case "eqfield":
		return "Passwords do not match"
	}
	fieldNameToUse := err.Field()
	if strings.Contains(fieldNameToUse, "ID") {
		fieldNameToUse = strings.ReplaceAll(fieldNameToUse, "ID", "")
	}
	return fmt.Sprintf("%s is %s", fieldNameToUse, err.Tag())
}

func ToErrors(errors any) []SimpleValidation {
	errs := make([]SimpleValidation, 0)
	switch v := errors.(type) {
	case string:
		if len(v) > 0 {
			return []SimpleValidation{
				{
					Message: v,
				},
			}
		}
	case validator.ValidationErrors:
		for _, vErr := range v {
			errs = append(errs, SimpleValidation{
				Message: friendlyError(vErr),
			})
		}
	case []SimpleValidation:
		return v
	}

	return errs
}
