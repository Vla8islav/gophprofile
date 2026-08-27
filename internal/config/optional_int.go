package config

import (
	"fmt"
	"strconv"
)

type OptionalInt struct {
	Value   int
	BeenSet bool
}

func (b *OptionalInt) String() string {
	return strconv.Itoa(b.Value)
}

func (b *OptionalInt) Set(s string) error {
	val, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	b.Value = val
	b.BeenSet = true
	return nil
}

func (b *OptionalInt) UnmarshalText(text []byte) error {
	v := string(text)
	result, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("cannot convert string to int: %w", err)
	}
	b.Value = result
	b.BeenSet = true
	return nil
}
