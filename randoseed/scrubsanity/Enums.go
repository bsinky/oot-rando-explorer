package scrubsanity

type Scrubsanity int

const (
	Off Scrubsanity = iota
	Affordable
	Expensive
	RandomPrices
)

func (v Scrubsanity) DisplayName() string {
	if v == RandomPrices {
		return "Random Prices"
	}
	return v.String()
}

func FromDisplayName(displayName string) (Scrubsanity, error) {
	return ScrubsanityString(displayName)
}

type ScrubsanityEnum struct{}

func (e ScrubsanityEnum) Values() []int {
	ret := make([]int, 0)
	for _, v := range ScrubsanityValues() {
		ret = append(ret, int(v))
	}
	return ret
}

func (e ScrubsanityEnum) ValueMap() map[int]string {
	vals := ScrubsanityValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = Scrubsanity(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=Scrubsanity
