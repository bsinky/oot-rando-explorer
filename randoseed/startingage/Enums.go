package startingage

import (
	"errors"
	"strings"
)

type StartingAge int

const (
	Child StartingAge = iota
	Adult
	Random
)

func (v StartingAge) DisplayName() string {
	switch v {
	case Child:
		return "Child"
	case Adult:
		return "Adult"
	case Random:
		return "Random"
	default:
		return v.String()
	}
}

func FromDisplayName(name string) (StartingAge, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return Random, nil
	}
	switch strings.ToLower(n) {
	case "child":
		return Child, nil
	case "adult":
		return Adult, nil
	case "random":
		return Random, nil
	}
	return Random, errors.New("unknown starting age: " + name)
}

type StartingAgeEnum struct{}

func (e StartingAgeEnum) ValueMap() map[int]string {
	vals := StartingAgeValues()
	m := make(map[int]string, len(vals))
	for _, v := range vals {
		m[int(v)] = v.String()
	}
	return m
}

//go:generate go run github.com/dmarkham/enumer -type=StartingAge
