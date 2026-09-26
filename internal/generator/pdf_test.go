package generator

import (
	"strings"
	"testing"

	"guias-despacho/internal/domain"
)

func TestGenerarPDF(t *testing.T) {
	pdf, err := GenerarPDF(domain.Ejemplo())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF") {
		t.Fatal("el archivo no es un PDF")
	}
	if len(pdf) < 1500 {
		t.Fatalf("PDF demasiado pequeño: %d bytes", len(pdf))
	}
}
