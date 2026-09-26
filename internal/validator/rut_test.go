package validator

import "testing"

func TestDigitoVerificador(t *testing.T) {
	casos := []struct {
		cuerpo string
		dv     string
	}{
		{"11111111", "1"},
		{"12345678", "5"},
		{"76123456", "0"},
		{"76543210", "3"},
		{"99999999", "9"},
		{"15678932", "1"},
		{"100000", "4"},
	}
	for _, c := range casos {
		if got := DigitoVerificador(c.cuerpo); got != c.dv {
			t.Errorf("DV(%s) = %s, quiere %s", c.cuerpo, got, c.dv)
		}
	}
}

func TestRutValido(t *testing.T) {
	validos := []string{"11111111-1", "12.345.678-5", "76123456-0", "100.008-k", "100000-4"}
	for _, rut := range validos {
		if !RutValido(NormalizarRUT(rut)) {
			t.Errorf("se esperaba RUT válido: %s", rut)
		}
	}
	invalidos := []string{"12345678-9", "11111111-K", "123", "12.345.678-4", "00123456-5", ""}
	for _, rut := range invalidos {
		if RutValido(NormalizarRUT(rut)) {
			t.Errorf("se esperaba RUT inválido: %s", rut)
		}
	}
}
