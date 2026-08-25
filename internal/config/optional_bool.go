package config

import "strconv"

type OptionalBool struct {
	Value   bool
	BeenSet bool
}

func (b *OptionalBool) String() string {
	return strconv.FormatBool(b.Value)
}

func (b *OptionalBool) Set(s string) error {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	b.Value = val
	b.BeenSet = true
	return nil
}

func (b *OptionalBool) UnmarshalText(text []byte) error {
	v, err := strconv.ParseBool(string(text))
	if err != nil {
		return err
	}
	b.Value = v
	b.BeenSet = true
	return nil
}
