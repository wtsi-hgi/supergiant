package common

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ID is defined as a string pointer in order to check for nil in the context of
// relations. NOTE this may not be best practice.
type ID *string

type Timestamp struct {
	time.Time
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Time.Format(time.RFC1123))
	return []byte(stamp), nil
}

func (t *Timestamp) UnmarshalJSON(raw []byte) (err error) {
	str := string(raw)
	noQuotes := str[1 : len(str)-1]
	t.Time, err = time.Parse(time.RFC1123, noQuotes)
	return
}

func NewTimestamp() *Timestamp {
	return &Timestamp{time.Now().UTC()}
}

func TimestampFromString(str string) *Timestamp {
	t := new(Timestamp)
	t.UnmarshalJSON([]byte(fmt.Sprintf(`"%s"`, str)))
	return t
}

// CPU / RAM resource values

type BytesValue struct {
	Bytes int64
}

const (
	kibibytes int64 = 1024
	mebibytes       = kibibytes * kibibytes
	gibibytes       = mebibytes * kibibytes
)

func (b *BytesValue) MarshalJSON() ([]byte, error) {
	return []byte(`"` + b.ToKubeMebibytes() + `"`), nil
}

func (b *BytesValue) UnmarshalJSON(raw []byte) error {
	str := string(raw)

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

func (c *CoresValue) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.ToKubeMillicores() + `"`), nil
}

func (c *CoresValue) UnmarshalJSON(raw []byte) error {
	str := string(raw)

	rxpMillicores := regexp.MustCompile(`^"([0-9]+)m"$`)        // 1000m
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

// func (v *CoresValue) cores() float64 {
// 	return float64(v.millicores) / float64(millicores)
// }

func (v *CoresValue) ToKubeMillicores() string {
	return fmt.Sprintf("%dm", v.Millicores)
}

func (v *CoresValue) Cores() float64 {
	return float64(millicores) / float64(millicores)
}
