package configuration

import (
	"time"

	"github.com/go-errors/errors"
)

type KeyMap map[string][]string

type Configuration interface {
	// All returns a map of all flattened key paths and their values.
	// Note that it uses maps.Copy to create a copy that uses
	// json.Marshal which changes the numeric types to float64.
	All() map[string]interface{}

	// Bool returns the bool value of a given key path or false if the path
	// does not exist or if the value is not a valid bool representation.
	// Accepted string representations of bool are the ones supported by strconv.ParseBool.
	Bool(path string) bool

	// Bools returns the []bool slice value of a given key path or an
	// empty []bool slice if the path does not exist or if the value
	// is not a valid bool slice.
	Bools(path string) []bool

	// MustBools returns the []bool value of a given key path or panics
	// if the value is not set or set to default value.
	MustBools(path string) []bool

	// BoolMap returns the map[string]bool value of a given key path
	// or an empty map[string]bool if the path does not exist or if the
	// value is not a valid bool map.
	BoolMap(path string) map[string]bool

	// MustBoolMap returns the map[string]bool value of a given key path or panics
	// if the value is not set or set to default value.
	MustBoolMap(path string) map[string]bool

	// Bytes returns the []byte value of a given key path or an empty
	// []byte slice if the path does not exist or if the value is not a valid string.
	Bytes(path string) []byte

	// MustBytes returns the []byte value of a given key path or panics
	// if the value is not set or set to default value.
	MustBytes(path string) []byte

	// Delim returns delimiter in used by this instance of Koanf.
	Delim() string

	// Exists returns true if the given key path exists in the conf map.
	Exists(path string) bool

	// Float64 returns the float64 value of a given key path or 0 if the path
	// does not exist or if the value is not a valid float64.
	Float64(path string) float64

	// MustFloat64 returns the float64 value of a given key path or panics
	// if it isn't set or set to default value 0.
	MustFloat64(path string) float64

	// Float64s returns the []float64 slice value of a given key path or an
	// empty []float64 slice if the path does not exist or if the value
	// is not a valid float64 slice.
	Float64s(path string) []float64

	// MustFloat64s returns the []Float64 slice value of a given key path or panics
	// if the value is not set or set to default value.
	MustFloat64s(path string) []float64

	// Float64Map returns the map[string]float64 value of a given key path
	// or an empty map[string]float64 if the path does not exist or if the
	// value is not a valid float64 map.
	Float64Map(path string) map[string]float64

	// MustFloat64Map returns the map[string]float64 value of a given key path or panics
	// if the value is not set or set to default value.
	MustFloat64Map(path string) map[string]float64

	// Get returns the raw, uncast interface{} value of a given key path
	// in the config map. If the key path does not exist, nil is returned.
	Get(path string) interface{}

	// Duration returns the time.Duration value of a given key path assuming
	// that the key contains a valid numeric value.
	Duration(path string) time.Duration

	// MustDuration returns the time.Duration value of a given key path or panics
	// if it isn't set or set to default value 0.
	MustDuration(path string) time.Duration

	// Int returns the int value of a given key path or 0 if the path
	// does not exist or if the value is not a valid int.
	Int(path string) int

	// MustInt returns the int value of a given key path or panics
	// if it isn't set or set to default value of 0.
	MustInt(path string) int

	// Ints returns the []int slice value of a given key path or an
	// empty []int slice if the path does not exist or if the value
	// is not a valid int slice.
	Ints(path string) []int

	// MustInts returns the []int slice value of a given key path or panics
	// if the value is not set or set to default value.
	MustInts(path string) []int

	// IntMap returns the map[string]int value of a given key path
	// or an empty map[string]int if the path does not exist or if the
	// value is not a valid int map.
	IntMap(path string) map[string]int

	// MustIntMap returns the map[string]int value of a given key path or panics
	// if the value is not set or set to default value.
	MustIntMap(path string) map[string]int

	// Int64 returns the int64 value of a given key path or 0 if the path
	// does not exist or if the value is not a valid int64.
	Int64(path string) int64

	// MustInt64 returns the int64 value of a given key path or panics
	// if the value is not set or set to default value of 0.
	MustInt64(path string) int64

	// Int64s returns the []int64 slice value of a given key path or an
	// empty []int64 slice if the path does not exist or if the value
	// is not a valid int slice.
	Int64s(path string) []int64

	// MustInt64s returns the []int64 slice value of a given key path or panics
	// if the value is not set or its default value.
	MustInt64s(path string) []int64

	// Int64Map returns the map[string]int64 value of a given key path
	// or an empty map[string]int64 if the path does not exist or if the
	// value is not a valid int64 map.
	Int64Map(path string) map[string]int64

	// MustInt64Map returns the map[string]int64 value of a given key path
	// or panics if it isn't set or set to default value.
	MustInt64Map(path string) map[string]int64

	// Keys returns the slice of all flattened keys in the loaded configuration
	// sorted alphabetically.
	Keys() []string

	// MapKeys returns a sorted string list of keys in a map addressed by the
	// given path. If the path is not a map, an empty string slice is
	// returned.
	MapKeys(path string) []string

	// Raw returns a copy of the full raw conf map.
	// Note that it uses maps.Copy to create a copy that uses
	// json.Marshal which changes the numeric types to float64.
	Raw() map[string]interface{}

	// String returns the string value of a given key path or "" if the path
	// does not exist or if the value is not a valid string.
	String(path string) string

	// MustString returns the string value of a given key path
	// or panics if it isn't set or set to default value "".
	MustString(path string) string

	// Strings returns the []string slice value of a given key path or an
	// empty []string slice if the path does not exist or if the value
	// is not a valid string slice.
	Strings(path string) []string

	// MustStrings returns the []string slice value of a given key path or panics
	// if the value is not set or set to default value.
	MustStrings(path string) []string

	// StringMap returns the map[string]string value of a given key path
	// or an empty map[string]string if the path does not exist or if the
	// value is not a valid string map.
	StringMap(path string) map[string]string

	// MustStringMap returns the map[string]string value of a given key path or panics
	// if the value is not set or set to default value.
	MustStringMap(path string) map[string]string

	// StringsMap returns the map[string][]string value of a given key path
	// or an empty map[string][]string if the path does not exist or if the
	// value is not a valid strings map.
	StringsMap(path string) map[string][]string

	// MustStringsMap returns the map[string][]string value of a given key path or panics
	// if the value is not set or set to default value.
	MustStringsMap(path string) map[string][]string

	// Time attempts to parse the value of a given key path and return time.Time
	// representation. If the value is numeric, it is treated as a UNIX timestamp
	// and if it's string, a parse is attempted with the given layout.
	Time(path, layout string) time.Time

	// MustTime attempts to parse the value of a given key path and return time.Time
	// representation. If the value is numeric, it is treated as a UNIX timestamp
	// and if it's string, a parse is attempted with the given layout. It panics if
	// the parsed time is zero.
	MustTime(path, layout string) time.Time
}

// ErrHelp may be returned by ConfigProvider invocation to indicate that the user specified `--help` when invoking the program.
var ErrHelp = errors.New("flags: help requested")

// Metadata is a definition of a single configuration. Every configuration must have a name, an optional default value, and a usage
// description.
type Metadata struct {
	Name         string
	DefaultValue string
	Usage        string
}

// ConfigProvider is a function that accepts a list of configuration metadata--definition of parameters that will be used by the program--
// and returns a Configuration instance.
//
// ErrHelp may be returned if the user specified `--help` when invoking the program.
type ConfigProvider func(stringConfigs []Metadata, boolConfigs []Metadata) (Configuration, error)
