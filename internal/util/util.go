package util

import "encoding/json"

// ToInt converts any built-in numeric type to int.
// Returns false if v is not a recognized numeric type.
func ToInt(v any) (int, bool) {
	switch n := v.(type) {
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
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
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
