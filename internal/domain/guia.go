package domain

import (
	"math"
	"strconv"
	"strings"
)

// TipoTraslado es un valor de IndTraslado del formato DTE.
// El código 7 usa la glosa vigente del Anexo Técnico 2.5: "Devolución de mercaderías".
type TipoTraslado struct {
	Codigo int
	Glosa  string
}

// TiposTraslado es el catálogo del indicador de traslado de bienes.
func TiposTraslado() []TipoTraslado {
	return []TipoTraslado{
		{1, "Operación constituye venta"},
		{2, "Ventas por efectuar"},
		{3, "Consignaciones"},
		{4, "Entrega gratuita"},
		{5, "Traslados internos"},
		{6, "Otros traslados no venta"},
		{7, "Devolución de mercaderías"},
		{8, "Traslado para exportación"},
		{9, "Venta para exportación"},
	}
}

// GlosaTraslado devuelve la glosa del indicador o una etiqueta genérica.
func GlosaTraslado(codigo int) string {
	for _, t := range TiposTraslado() {
		if t.Codigo == codigo {
			return t.Glosa
		}
	}
	return "Traslado no clasificado"
}

// GlosaIndExe describe el tratamiento tributario de la línea.
func GlosaIndExe(codigo int) string {
	switch codigo {
	case 1:
		return "Exento"
	case 2:
		return "No facturable"
	default:
		return "Afecto"
	}
}

// Item es una línea de mercadería de la guía.
type Item struct {
	NmbItem        string
	QtyItem        float64
	UnmdItem       string
	PrcItem        float64
	MontoItem      int64
	IndExe         int
	MontoInformado bool
	MontoInvalido  bool
}

// Guia es el modelo de trabajo de una Guía de Despacho Electrónica (DTE 52).
// No representa un DTE firmado: no incluye CAF, TED ni firma digital.
type Guia struct {
	Folio       int64
	FchEmis     string
	IndTraslado int

	RUTEmisor string
	RznSoc    string
	GiroEmis  string
	Acteco    string

	RUTRecep    string
	RznSocRecep string
	GiroRecep   string
	DirRecep    string
	CmnaRecep   string

	RUTChofer    string
	NombreChofer string

	Patente          string
	PatenteCarro     string
	TransportePropio bool
	RUTTrans         string

	FchSalida      string
	HraSalida      string
	FchLlegada     string
	Multidia       bool
	MotivoMultidia string

	DirOrigen    string
	CmnaOrigen   string
	CiudadOrigen string
	DirDest      string
	CmnaDest     string
	CiudadDest   string

	Items []Item
}

// Totales resume los montos de la guía. Lo no facturable queda fuera del total.
type Totales struct {
	MntNeto         int64
	MntExe          int64
	IVA             int64
	MntTotal        int64
	MntNoFacturable int64
	TasaIVA         int
}

// Totales calcula neto, exento, IVA (19 %) y total a partir del detalle.
func (g Guia) Totales() Totales {
	var t Totales
	for _, it := range g.Items {
		switch it.IndExe {
		case 1:
			t.MntExe += it.MontoItem
		case 2:
			t.MntNoFacturable += it.MontoItem
		default:
			t.MntNeto += it.MontoItem
		}
	}
	if t.MntNeto > 0 {
		t.TasaIVA = 19
		t.IVA = int64(math.Round(float64(t.MntNeto) * 0.19))
	}
	t.MntTotal = t.MntNeto + t.IVA + t.MntExe
	return t
}

// FormatoPesos presenta un monto en pesos chilenos, sin decimales.
func FormatoPesos(n int64) string {
	signo := ""
	if n < 0 {
		signo = "-"
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	b.WriteString(signo)
	b.WriteByte('$')
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// FormatoCantidad muestra la cantidad con coma decimal, sin ceros de relleno.
func FormatoCantidad(n float64) string {
	s := strconv.FormatFloat(n, 'f', -1, 64)
	return strings.ReplaceAll(s, ".", ",")
}

// Ejemplo es una guía que cumple las reglas de la Res. Ex. N° 154
// y sirve para pruebas y para el XML de referencia.
func Ejemplo() Guia {
	return Guia{
		Folio:            152,
		FchEmis:          "2026-11-02",
		IndTraslado:      1,
		RUTEmisor:        "76123456-0",
		RznSoc:           "Molinos del Valle SpA",
		GiroEmis:         "Elaboración de productos de molinería",
		Acteco:           "106101",
		RUTRecep:         "76543210-3",
		RznSocRecep:      "Panaderia Sur Ltda",
		GiroRecep:        "Elaboración de pan",
		DirRecep:         "Av. Costanera 450",
		CmnaRecep:        "Concepcion",
		RUTChofer:        "15678932-1",
		NombreChofer:     "JUAN PEDRO SOTO LAGOS",
		Patente:          "ABCD12",
		PatenteCarro:     "XY1234",
		TransportePropio: false,
		RUTTrans:         "99999999-9",
		FchSalida:        "2026-11-02",
		HraSalida:        "08:30:00",
		FchLlegada:       "2026-11-03",
		Multidia:         true,
		MotivoMultidia:   "Traslado interregional con pernocta en ruta",
		DirOrigen:        "Camino Industrial 1200",
		CmnaOrigen:       "Pudahuel",
		CiudadOrigen:     "Santiago",
		DirDest:          "Av. Costanera 450",
		CmnaDest:         "Concepcion",
		CiudadDest:       "Concepcion",
		Items: []Item{
			{
				NmbItem:        "Harina de trigo",
				QtyItem:        10,
				UnmdItem:       "KG",
				PrcItem:        1200,
				MontoItem:      12000,
				IndExe:         0,
				MontoInformado: true,
			},
			{
				NmbItem:        "Sacos de papel",
				QtyItem:        5,
				UnmdItem:       "UN",
				PrcItem:        0,
				MontoItem:      0,
				IndExe:         1,
				MontoInformado: true,
			},
		},
	}
}
