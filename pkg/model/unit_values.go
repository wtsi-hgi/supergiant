package model

import (
	"fmt"
	"regexp"
	"strconv"
)

// CPU / RAM resource values

type BytesValue struct {
	Bytes int64
}

const (
	kibibytes int64 = 1024
	mebibytes       = kibibytes * kibibytes
	gibibytes       = mebibytes * kibibytes
)

func BytesFromString(str string) *BytesValue {
	b := new(BytesValue)
	b.fromString(str) // NOTE error ignored
	return b
}

func (b *BytesValue) fromString(str string) error {
	rxp := regexp.MustCompile(`^"?([0-9]+(?:\.[0-9]+)?)([KMG]i)?"?$`)

	if !rxp.MatchString(str) {
		return fmt.Errorf(`Bytes value %s does not match regex ^"?([0-9]+(?:\.[0-9]+)?)([KMG]i)?"?$`, str)
	}

	match := rxp.FindStringSubmatch(str)

	float, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return err
	}

	switch match[2] {
	case "":
		b.Bytes = int64(float)
	case "Mi":
		b.Bytes = int64(float * float64(mebibytes))
	case "Gi":
		b.Bytes = int64(float * float64(gibibytes))
	default:
		return fmt.Errorf("Cannot parse bytes value from %s", str)
	}

	return nil
}

func (b *BytesValue) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.ToKubeMebibytes() + `"`), nil
}

func (b *BytesValue) UnmarshalJSON(raw []byte) error {
	return b.fromString(string(raw))
}

func (v *BytesValue) Mebibytes() float64 {
	return float64(v.Bytes) / float64(mebibytes)
}

func (v *BytesValue) Gibibytes() float64 {
	return float64(v.Bytes) / float64(gibibytes)
}

func (v *BytesValue) ToKubeMebibytes() string {
	return fmt.Sprintf("%dMi", int(v.Mebibytes()))
}

type CoresValue struct {
	Millicores int
}

const millicores = 1000

func CoresFromString(str string) *CoresValue {
	c := new(CoresValue)
	c.fromString(str)
	return c
}

func (c *CoresValue) fromString(str string) error {
	rxpMillicores := regexp.MustCompile(`^"?([0-9]+)m"?$`)      // 1000m
	rxpCores := regexp.MustCompile(`^"?([0-9]+(\.[0-9]+)?)"?$`) // 1 (can have quotes)

	getNumMatch := func(rxp *regexp.Regexp) (float64, error) {
		numberStr := rxp.FindStringSubmatch(str)[1]
		return strconv.ParseFloat(numberStr, 64)
	}

	switch {
	case rxpMillicores.MatchString(str):

		num, err := getNumMatch(rxpMillicores)
		if err != nil {
			return err
		}
		c.Millicores = int(num)

	case rxpCores.MatchString(str):

		num, err := getNumMatch(rxpCores)
		if err != nil {
			return err
		}
		c.Millicores = int(num * millicores)

	default:
		return fmt.Errorf("Could not parse cores value from %s", str)
	}
	return nil
}

func (c *CoresValue) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.ToKubeMillicores() + `"`), nil
}

func (c *CoresValue) UnmarshalJSON(raw []byte) error {
	return c.fromString(string(raw))
}

func (v *CoresValue) ToKubeMillicores() string {
	return fmt.Sprintf("%dm", v.Millicores)
}

func (v *CoresValue) Cores() float64 {
	return float64(v.Millicores) / float64(millicores)
}
