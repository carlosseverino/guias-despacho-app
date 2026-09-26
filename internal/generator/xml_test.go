package generator

import (
	"strings"
	"testing"

	"guias-despacho/internal/domain"
)

func TestGenerarXMLTipo52(t *testing.T) {
	xml, err := GenerarXML(domain.Ejemplo())
	if err != nil {
		t.Fatal(err)
	}
	texto := string(xml)
	if !strings.Contains(texto, `version="1.0"`) {
		t.Fatal("el atributo version del DTE debe permanecer en 1.0")
	}
	if strings.Contains(texto, "<TED>") || strings.Contains(texto, "<Signature") {
		t.Fatal("el borrador no debe incluir timbre ni firma")
	}
	for _, tag := range []string{
		"<TipoDTE>52</TipoDTE>",
		"<RUTEmisor>76123456-0</RUTEmisor>",
		"<RUTChofer>15678932-1</RUTChofer>",
		"<NombreChofer>JUAN PEDRO SOTO LAGOS</NombreChofer>",
		"<Patente>ABCD12</Patente>",
		"<PatenteCarro>XY1234</PatenteCarro>",
		"<RUTTrans>99999999-9</RUTTrans>",
		"<DirOrigen>Camino Industrial 1200</DirOrigen>",
		"<CmnaOrigen>Pudahuel</CmnaOrigen>",
		"<DirDest>Av. Costanera 450</DirDest>",
		"<CmnaDest>Concepcion</CmnaDest>",
		"<FchSalida>2026-11-02</FchSalida>",
		"<HraSalida>08:30:00</HraSalida>",
		"<FchLlegada>2026-11-03</FchLlegada>",
		"<IndExe>1</IndExe>",
		"<MntNeto>12000</MntNeto>",
		"<IVA>2280</IVA>",
		"<MntTotal>14280</MntTotal>",
	} {
		if !strings.Contains(texto, tag) {
			t.Errorf("falta %s", tag)
		}
	}

	orden := []string{
		"<Patente>", "<PatenteCarro>", "<RUTTrans>", "<Chofer>",
		"<RUTChofer>", "<NombreChofer>", "</Chofer>",
		"<DirDest>", "<CmnaDest>", "<CiudadDest>",
		"<FchSalida>", "<HraSalida>", "<FchLlegada>",
	}
	pos := -1
	for _, tag := range orden {
		i := strings.Index(texto, tag)
		if i < 0 || i < pos {
			t.Fatalf("orden de Transporte incorrecto en %s", tag)
		}
		pos = i
	}
}

func TestGenerarXMLRechazaInvalida(t *testing.T) {
	g := domain.Ejemplo()
	g.HraSalida = ""
	if _, err := GenerarXML(g); err == nil {
		t.Fatal("debía rechazar la guía sin hora de salida")
	}
}
