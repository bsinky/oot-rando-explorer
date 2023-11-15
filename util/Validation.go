package util

type SimpleValidation struct {
	Message string
}

func (u *SimpleValidation) Error() string {
	return u.Message
}

func (u *SimpleValidation) FieldName() string {
	return ""
}
