package util

import "github.com/go-playground/validator/v10"

type SimpleValidation struct {
	Message string
}

func (u *SimpleValidation) Error() string {
	return u.Message
}

func (u *SimpleValidation) FieldName() string {
	return ""
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
				Message: vErr.Error(),
			})
		}
	}

	return errs
}
