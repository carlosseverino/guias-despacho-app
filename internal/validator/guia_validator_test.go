package validator

import (
	"strings"
	"testing"

	"guias-despacho/internal/domain"
)

func TestEjemploCumpleResolucion154(t *testing.T) {
	if errs := ValidarGuia(domain.Ejemplo()); len(errs) != 0 {
		t.Fatalf("la guía de ejemplo no cumple: %+v", errs)
	}
}

func TestReglasResolucion154(t *testing.T) {
	casos := []struct {
		nombre string
		mutar  func(*domain.Guia)
		campo  string
	}{
		{"rut emisor", func(g *domain.Guia) { g.RUTEmisor = "11111111-K" }, "RUTEmisor"},
		{"rut chofer", func(g *domain.Guia) { g.RUTChofer = "" }, "RUTChofer"},
		{"nombre chofer", func(g *domain.Guia) { g.NombreChofer = "" }, "NombreChofer"},
		{"patente", func(g *domain.Guia) { g.Patente = "ABC123" }, "Patente"},
		{"patente carro", func(g *domain.Guia) { g.PatenteCarro = "12AB" }, "PatenteCarro"},
		{"tercero sin rut", func(g *domain.Guia) { g.RUTTrans = "" }, "RUTTrans"},
		{"tercero igual al emisor", func(g *domain.Guia) { g.RUTTrans = g.RUTEmisor }, "RUTTrans"},
		{"hora vacía", func(g *domain.Guia) { g.HraSalida = "" }, "HraSalida"},
		{"hora inválida", func(g *domain.Guia) { g.HraSalida = "25:00:00" }, "HraSalida"},
		{"multidia sin llegada", func(g *domain.Guia) { g.FchLlegada = "" }, "FchLlegada"},
		{"multidia sin motivo", func(g *domain.Guia) { g.MotivoMultidia = "" }, "MotivoMultidia"},
		{"sin origen", func(g *domain.Guia) { g.DirOrigen = "" }, "DirOrigen"},
		{"sin destino", func(g *domain.Guia) { g.CmnaDest = "" }, "CmnaDest"},
		{"item sin monto", func(g *domain.Guia) {
			g.Items = []domain.Item{{NmbItem: "Trigo", QtyItem: 1, UnmdItem: "KG", IndExe: 0}}
		}, "Items[0].MontoItem"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			g := domain.Ejemplo()
			c.mutar(&g)
			if !tieneCampo(ValidarGuia(g), c.campo) {
				t.Fatalf("se esperaba error en %s", c.campo)
			}
		})
	}
}

func TestMismoDiaNoExigeLlegada(t *testing.T) {
	g := domain.Ejemplo()
	g.Multidia = false
	g.FchLlegada = ""
	g.MotivoMultidia = ""
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("un traslado del mismo día no debe exigir FchLlegada: %+v", errs)
	}
}

func TestLlegadaPosteriorExigeMarcaMultidia(t *testing.T) {
	g := domain.Ejemplo()
	g.Multidia = false
	g.MotivoMultidia = ""
	if !tieneCampo(ValidarGuia(g), "FchLlegada") {
		t.Fatal("se esperaba exigir la marca de más de un día")
	}
}

func TestExentoSinMonto(t *testing.T) {
	g := domain.Ejemplo()
	g.Items = []domain.Item{{
		NmbItem: "Muestra", QtyItem: 2, UnmdItem: "UN", IndExe: 2,
	}}
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("un ítem no facturable puede ir sin monto: %+v", errs)
	}
}

func TestPatenteCarroVacia(t *testing.T) {
	g := domain.Ejemplo()
	g.PatenteCarro = ""
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("PatenteCarro es opcional: %+v", errs)
	}
}

func TestTransportePropioSinRUTTrans(t *testing.T) {
	g := domain.Ejemplo()
	g.TransportePropio = true
	g.RUTTrans = ""
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("el transporte propio no exige RUTTrans: %+v", errs)
	}
}

func TestHoraSinSegundosSeCompleta(t *testing.T) {
	g := domain.Ejemplo()
	g.HraSalida = "08:30"
	NormalizarGuia(&g)
	if g.HraSalida != "08:30:00" {
		t.Fatalf("hora normalizada = %s", g.HraSalida)
	}
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("HH:MM debe aceptarse y completarse: %+v", errs)
	}
}

func TestMontoSeCalculaDesdePrecio(t *testing.T) {
	g := domain.Ejemplo()
	g.Items = []domain.Item{{
		NmbItem: "Aceite", QtyItem: 2.5, UnmdItem: "LT", PrcItem: 1000, IndExe: 0,
	}}
	if errs := ValidarGuia(g); len(errs) != 0 {
		t.Fatalf("debía calcular el monto: %+v", errs)
	}
	NormalizarGuia(&g)
	if g.Items[0].MontoItem != 2500 {
		t.Fatalf("monto = %d", g.Items[0].MontoItem)
	}
}

func tieneCampo(errs []Incumplimiento, campo string) bool {
	for _, e := range errs {
		if e.Campo == campo || strings.HasPrefix(e.Campo, campo) {
			return true
		}
	}
	return false
}
