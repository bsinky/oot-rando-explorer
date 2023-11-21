package tokensanity

type Tokensanity int

const (
	Off Tokensanity = iota
	Dungeons
	Overworld
	AllTokens
)

func (v Tokensanity) DisplayName() string {
	if v == AllTokens {
		return "All Tokens"
	}
	return v.String()
}

func FromDisplayName(displayName string) (Tokensanity, error) {
	if displayName == AllTokens.DisplayName() {
		return AllTokens, nil
	}
	return TokensanityString(displayName)
}

type TokensanityEnum struct{}

func (e TokensanityEnum) ValueMap() map[int]string {
	vals := TokensanityValues()
	m := make(map[int]string, len(vals))
	for v := range vals {
		m[int(v)] = Tokensanity(v).DisplayName()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=Tokensanity
