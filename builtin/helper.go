package builtin

import (
	"fmt"
	"strconv"

	"github.com/rendis/devtoolkit"
)

// TryToCastToInt tries to cast the given value to an int.
// If the value is already an int, it will return the value and true.
// If the value is a string, it will try to parse it to an int and return the result and true if successful.
// Otherwise, it will return 0 and false.
func TryToCastToInt(v any) (int, bool) {
	if m, ok := devtoolkit.ToInt(v); ok {
		return m, true
	}

	switch vt := v.(type) {
	case string:
		if s, err := strconv.Atoi(vt); err == nil {
			return s, true
		}
	}

	return 0, false
}

// GetKeyAsInt tries to get the value of the given key from the map and cast it to an int.
// If the key does not exist or the value cannot be cast to an int, it will return 0 and false.
func GetKeyAsInt(key string, md map[string]any) (int, bool) {
	if v, ok := md[key]; ok {
		return TryToCastToInt(v)
	}
	return 0, false
}

// TryToCastToString tries to cast the given value to a string.
// If the value is already a string, it will return the value and true.
// Otherwise, it will return "" and false.
func TryToCastToString(v any) (string, bool) {
	switch vt := v.(type) {
	case string:
		return vt, true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", vt), true
	case float32, float64:
		return fmt.Sprintf("%f", vt), true
	case bool:
		return fmt.Sprintf("%t", vt), true
	default:
		return "", false
	}
}

// GetKeyAsString tries to get the value of the given key from the map and cast it to a string.
// If the key does not exist or the value cannot be cast to a string, it will return "" and false.
func GetKeyAsString(key string, md map[string]any) (string, bool) {
	v, ok := md[key]
	if ok {
		return TryToCastToString(v)
	}
	return "", false
}
