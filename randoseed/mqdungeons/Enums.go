package mqdungeons

type MQDungeons int

const (
	Selection MQDungeons = iota
	Random
	Zero
	One
	Two
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Eleven
	Twelve
)

func (v MQDungeons) DisplayName() string {
	switch v {
	case Zero:
		return "0"
	case One:
		return "1"
	case Two:
		return "2"
	case Three:
		return "3"
	case Four:
		return "4"
	case Five:
		return "5"
	case Six:
		return "6"
	case Seven:
		return "7"
	case Eight:
		return "8"
	case Nine:
		return "9"
	case Ten:
		return "10"
	case Eleven:
		return "11"
	case Twelve:
		return "12"
	default:
		return MQDungeons(v).String()
	}
}

func FromDisplayName(displayName string) (MQDungeons, error) {
	for v := range MQDungeonsValues() {
		if displayName == MQDungeons(v).DisplayName() {
			return MQDungeons(v), nil
		}
	}
	return MQDungeonsString(displayName)
}

type MQDungeonsEnum struct{}

func (e MQDungeonsEnum) ValueMap() map[int]string {
	vals := MQDungeonsValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = MQDungeons(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=MQDungeons
