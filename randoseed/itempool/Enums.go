package itempool

type ItemPool int

const (
	Plentiful ItemPool = iota
	Balanced
	Scarce
	Minimal
)

func (v ItemPool) DisplayName() string {
	return v.String()
}

func FromDisplayName(displayName string) (ItemPool, error) {
	return ItemPoolString(displayName)
}

type ItemPoolEnum struct{}

func (e ItemPoolEnum) Values() []int {
	ret := make([]int, 0)
	for _, v := range ItemPoolValues() {
		ret = append(ret, int(v))
	}
	return ret
}

func (e ItemPoolEnum) ValueMap() map[int]string {
	vals := ItemPoolValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = ItemPool(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=ItemPool
