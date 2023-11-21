package shopsanity

type Shopsanity int

const (
	Off Shopsanity = iota
	ZeroItems
	OneItem
	TwoItems
	ThreeItems
	FourItems
	Random
)

func (v Shopsanity) DisplayName() string {
	switch v {
	case ZeroItems:
		return "0 Items"
	case OneItem:
		return "1 Item"
	case TwoItems:
		return "2 Items"
	case ThreeItems:
		return "3 Items"
	case FourItems:
		return "4 Items"
	default:
		return v.String()
	}
}

func FromDisplayName(displayName string) (Shopsanity, error) {
	for v := range ShopsanityValues() {
		if displayName == Shopsanity(v).DisplayName() {
			return Shopsanity(v), nil
		}
	}
	return ShopsanityString(displayName)
}

type ShopsanityEnum struct{}

func (e ShopsanityEnum) ValueMap() map[int]string {
	vals := ShopsanityValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = Shopsanity(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=Shopsanity
