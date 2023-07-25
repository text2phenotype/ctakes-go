package ml

import (
	"strings"
)

type Feature interface {
	String() string
}

type IntFeature struct {
	Name  string
	Value int
}

func (f *IntFeature) String() string {
	return f.Name
}

type StrFeature struct {
	Name  string
	Value string
}

func (f *StrFeature) String() string {
	parts := []string{f.Name, f.Value}
	return strings.Join(parts, "_")
}

type BoolFeature struct {
	Name  string
	Value bool
}

func (f *BoolFeature) String() string {
	return f.Name
}
