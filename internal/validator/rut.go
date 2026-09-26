package validator

import (
	"regexp"
	"strconv"
	"strings"
)

// reRUT exige cuerpo entre 100.000 y 99.999.999, guion y dígito 0-9 o K.
var reRUT = regexp.MustCompile(`^[1-9]\d{5,7}-[0-9K]$`)

// NormalizarRUT quita puntos y espacios, lleva el dígito a mayúscula
// e inserta el guion cuando el valor viene compacto (123456785).
func NormalizarRUT(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.NewReplacer(".", "", " ", "", "–", "-", "—", "-").Replace(s)
	if s == "" {
		return ""
	}
	if !strings.Contains(s, "-") && len(s) >= 2 {
		s = s[:len(s)-1] + "-" + s[len(s)-1:]
	}
	return s
}

// DigitoVerificador aplica el módulo 11 chileno.
// Desde la derecha, cada dígito se multiplica por 2, 3, 4, 5, 6, 7 y el ciclo se reinicia.
// El dígito es 11 menos el resto de la suma; 11 se traduce en 0 y 10 en K.
func DigitoVerificador(cuerpo string) string {
	suma := 0
	factor := 2
	for i := len(cuerpo) - 1; i >= 0; i-- {
		suma += int(cuerpo[i]-'0') * factor
		factor++
		if factor > 7 {
			factor = 2
		}
	}
	resto := 11 - (suma % 11)
	switch resto {
	case 11:
		return "0"
	case 10:
		return "K"
	default:
		return strconv.Itoa(resto)
	}
}

// RutValido comprueba sintaxis y dígito verificador.
// El RUT debe estar normalizado (sin puntos, con guion y K mayúscula).
func RutValido(rut string) bool {
	if !reRUT.MatchString(rut) {
		return false
	}
	cuerpo, dv, _ := strings.Cut(rut, "-")
	return DigitoVerificador(cuerpo) == dv
}
