package logic

type Logic int

const (
	Glitchless Logic = iota
	Glitched
	NoLogic
	Vanilla
)

type LogicEnum struct{}

func (v Logic) DisplayName() string {
	if v == NoLogic {
		return "No Logic"
	}
	return v.String()
}

func FromDisplayName(displayName string) (Logic, error) {
	if displayName == NoLogic.DisplayName() {
		return NoLogic, nil
	}
	return LogicString(displayName)
}

func (e LogicEnum) ValueMap() map[int]string {
	vals := LogicValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = Logic(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=Logic
