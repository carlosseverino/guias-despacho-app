package validator

import "testing"

func TestPatenteValida(t *testing.T) {
	validas := []string{"AB1234", "ab-1234", "ABCD12", "abcd-12", "XY 1234"}
	for _, p := range validas {
		if !PatenteValida(p) {
			t.Errorf("se esperaba patente válida: %s", p)
		}
	}
	invalidas := []string{"", "ABC123", "A12345", "AB12345", "ABCD123", "12AB34", "ABCD"}
	for _, p := range invalidas {
		if PatenteValida(p) {
			t.Errorf("se esperaba patente inválida: %s", p)
		}
	}
}
