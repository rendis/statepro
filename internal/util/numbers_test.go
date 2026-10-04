package util

import (
	"encoding/json"
	"testing"
)

func TestNormalizeJSONNumbers_ConservaValorDecimal(t *testing.T) {
	for _, test := range []struct {
		number  string
		precise bool
	}{
		{"42", true}, {"1.25", true}, {"1e3", true}, {"-0", true}, {"0e99999999999999999999", false},
		{"9007199254740993", false}, {"9007199254740994", true}, {"0.1234567890123456789", false}, {"1e400", false},
	} {
		t.Run(test.number, func(t *testing.T) {
			got := NormalizeJSONNumbers(json.Number(test.number))
			_, isFloat := got.(float64)
			if isFloat != test.precise {
				t.Fatalf("tipo inesperado: %#v", got)
			}
			if !test.precise && got != json.Number(test.number) {
				t.Fatal("numero alterado")
			}
		})
	}
}

func TestToInt_AceptaJSONNumberYRechazaOverflow(t *testing.T) {
	if value, ok := ToInt(json.Number("42")); !ok || value != 42 {
		t.Fatal("contador JSON no soportado")
	}
	if _, ok := ToInt(json.Number("18446744073709551615")); ok {
		t.Fatal("overflow aceptado")
	}
}
