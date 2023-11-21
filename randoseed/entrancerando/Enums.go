package entrancerando

type EntranceRando int

const (
	Off EntranceRando = iota
	On
)

func (v EntranceRando) DisplayName() string {
	return v.String()
}

func FromDisplayName(displayName string) (EntranceRando, error) {
	return EntranceRandoString(displayName)
}

type EntranceRandoEnum struct{}

func (e EntranceRandoEnum) ValueMap() map[int]string {
	vals := EntranceRandoValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = EntranceRando(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=EntranceRando
