package util

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// ToInt converts any built-in numeric type to int.
// Returns false if v is not a recognized numeric type.
func ToInt(v any) (int, bool) {
	switch n := v.(type) {
	case json.Number:
		value, err := strconv.Atoi(n.String())
		return value, err == nil
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case int32:
		return int(n), true
	case int16:
		return int(n), true
	case int8:
		return int(n), true
	case uint:
		return int(n), true
	case uint64:
		return int(n), true
	case uint32:
		return int(n), true
	case uint16:
		return int(n), true
	case uint8:
		return int(n), true
	default:
		return 0, false
	}
}

// StructToMap converts a struct to a map[string]any via JSON round-trip.
func StructToMap(s any) (map[string]any, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}
	NormalizeJSONNumbers(out)
	return out, nil
}

// NormalizeJSONNumbers keeps legacy float64 values when JSON round-tripping is
// exact, and retains json.Number when conversion would alter the numeric value.
// The maps and slices supplied by a JSON decoder are normalized in place.
func NormalizeJSONNumbers(value any) any {
	switch value := value.(type) {
	case map[string]any:
		for key, entry := range value {
			value[key] = NormalizeJSONNumbers(entry)
		}
	case []any:
		for index, entry := range value {
			value[index] = NormalizeJSONNumbers(entry)
		}
	case json.Number:
		candidate, err := strconv.ParseFloat(value.String(), 64)
		if err != nil {
			return value
		}
		encoded, err := json.Marshal(candidate)
		if err != nil {
			return value
		}
		original, ok := canonicalNumber(value.String())
		roundtrip, roundtripOK := canonicalNumber(string(encoded))
		if ok && roundtripOK && original == roundtrip {
			return candidate
		}
	}
	return value
}

type decimalNumber struct {
	negative bool
	digits   string
	exponent int64
}

// Compare decimal values without expanding exponents or allocating large integers.
func canonicalNumber(value string) (decimalNumber, bool) {
	var result decimalNumber
	if strings.HasPrefix(value, "-") {
		result.negative = true
		value = value[1:]
	}
	if index := strings.IndexAny(value, "eE"); index >= 0 {
		exponent, err := strconv.ParseInt(value[index+1:], 10, 64)
		if err != nil {
			return result, false
		}
		result.exponent = exponent
		value = value[:index]
	}
	if index := strings.IndexByte(value, '.'); index >= 0 {
		fractionLength := int64(len(value) - index - 1)
		if result.exponent < math.MinInt64+fractionLength {
			return result, false
		}
		result.exponent -= fractionLength
		value = value[:index] + value[index+1:]
	}
	value = strings.TrimLeft(value, "0")
	if value == "" {
		return decimalNumber{digits: "0"}, true
	}
	trimmed := strings.TrimRight(value, "0")
	trailing := int64(len(value) - len(trimmed))
	if result.exponent > math.MaxInt64-trailing {
		return result, false
	}
	result.exponent += trailing
	result.digits = trimmed
	return result, true
}

// MapToStructWithNumbers preserves numbers in interface-valued fields.
func MapToStructWithNumbers[T any](m map[string]any) (*T, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out T
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// MapToStruct converts a map[string]any to a struct of type T via JSON round-trip.
func MapToStruct[T any](m map[string]any) (*T, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Pair groups two values of arbitrary types.
type Pair[F any, S any] struct {
	First  F
	Second S
}

func NewPair[F any, S any](first F, second S) Pair[F, S] {
	return Pair[F, S]{First: first, Second: second}
}

func (p *Pair[F, S]) GetFirst() F    { return p.First }
func (p *Pair[F, S]) GetAll() (F, S) { return p.First, p.Second }
