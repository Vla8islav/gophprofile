package config

type OptionalString struct {
	Value   string
	BeenSet bool
}

func (b *OptionalString) String() string {
	return b.Value
}

func (b *OptionalString) Set(s string) error {
	b.Value = s
	b.BeenSet = true
	return nil
}

func (b *OptionalString) UnmarshalText(text []byte) error {
	v := string(text)
	b.Value = v
	b.BeenSet = true
	return nil
}
