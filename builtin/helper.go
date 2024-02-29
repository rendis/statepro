package builtin

import (
	"github.com/rendis/devtoolkit"
	"strconv"
)

func TryToCastToInt(v any) (int, bool) {
	if m, ok := devtoolkit.ToInt(v); ok {
		return m, true
	}

	switch v.(type) {
	case string:
		if s, err := strconv.Atoi(v.(string)); err == nil {
			return s, true
		}
	}

	return 0, false
}
