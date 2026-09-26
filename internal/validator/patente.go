package validator

import (
	"regexp"
	"strings"
)

// rePatente acepta la patente antigua (dos letras y cuatro dígitos)
// y la nueva (cuatro letras y dos dígitos).
var rePatente = regexp.MustCompile(`^(?:[A-Z]{2}\d{4}|[A-Z]{4}\d{2})$`)

// NormalizarPatente deja la patente en mayúsculas y sin separadores.
func NormalizarPatente(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	return strings.NewReplacer("-", "", " ", "", ".", "").Replace(s)
}

// PatenteValida informa si el valor, ya normalizado o no, cumple el formato chileno.
func PatenteValida(s string) bool {
	return rePatente.MatchString(NormalizarPatente(s))
}
