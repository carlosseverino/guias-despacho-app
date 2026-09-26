package generator

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"

	"guias-despacho/internal/domain"
	"guias-despacho/internal/validator"
)

// DocumentoInvalidoError indica que el XML o el PDF no se generaron
// porque la guía no pasó la prevalidación.
type DocumentoInvalidoError struct {
	Errores []validator.Incumplimiento
}

func (e DocumentoInvalidoError) Error() string {
	return "el documento no cumple las validaciones de la Res. Ex. N° 154"
}

// GenerarXML arma un DTE tipo 52 de trabajo, sin CAF, TED ni firma.
// El atributo version se mantiene en 1.0, como define el Anexo Técnico 2.5.
func GenerarXML(g domain.Guia) ([]byte, error) {
	validator.NormalizarGuia(&g)
	if errs := validator.ValidarGuia(g); len(errs) > 0 {
		return nil, DocumentoInvalidoError{Errores: errs}
	}
	return codificar(mapear(g))
}

func codificar(doc dteXML) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString("<!-- Borrador local estructurado según el Anexo Técnico DTE 2.5 y la Res. Ex. SII N°154. No incluye CAF, timbre electrónico ni firma digital, y no es un documento aceptado por el SII. -->\n")
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("codificar XML: %w", err)
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func mapear(g domain.Guia) dteXML {
	tot := g.Totales()
	return dteXML{
		Version: "1.0",
		Documento: documentoXML{
			ID: fmt.Sprintf("GD52F%d", g.Folio),
			Encabezado: encabezadoXML{
				IdDoc: idDocXML{
					TipoDTE:     52,
					Folio:       g.Folio,
					FchEmis:     g.FchEmis,
					IndTraslado: g.IndTraslado,
				},
				Emisor: emisorXML{
					RUTEmisor:    g.RUTEmisor,
					RznSoc:       g.RznSoc,
					GiroEmis:     g.GiroEmis,
					Acteco:       g.Acteco,
					DirOrigen:    g.DirOrigen,
					CmnaOrigen:   g.CmnaOrigen,
					CiudadOrigen: g.CiudadOrigen,
				},
				Receptor: receptorXML{
					RUTRecep:    g.RUTRecep,
					RznSocRecep: g.RznSocRecep,
					GiroRecep:   g.GiroRecep,
					DirRecep:    g.DirRecep,
					CmnaRecep:   g.CmnaRecep,
				},
				Transporte: transporteXML{
					Patente:      g.Patente,
					PatenteCarro: g.PatenteCarro,
					RUTTrans:     g.RUTTrans,
					Chofer: choferXML{
						RUTChofer:    g.RUTChofer,
						NombreChofer: g.NombreChofer,
					},
					DirDest:    g.DirDest,
					CmnaDest:   g.CmnaDest,
					CiudadDest: g.CiudadDest,
					FchSalida:  g.FchSalida,
					HraSalida:  g.HraSalida,
					FchLlegada: g.FchLlegada,
				},
				Totales: totalesXMLDesde(tot),
			},
			Detalle: detalleXMLDesde(g.Items),
		},
	}
}

func detalleXMLDesde(items []domain.Item) []detalleXML {
	out := make([]detalleXML, len(items))
	for i, it := range items {
		linea := detalleXML{
			NroLinDet: i + 1,
			IndExe:    it.IndExe,
			NmbItem:   it.NmbItem,
			QtyItem:   strconv.FormatFloat(it.QtyItem, 'f', -1, 64),
			UnmdItem:  it.UnmdItem,
			MontoItem: it.MontoItem,
		}
		if it.PrcItem > 0 {
			linea.PrcItem = strconv.FormatFloat(it.PrcItem, 'f', -1, 64)
		}
		out[i] = linea
	}
	return out
}

func totalesXMLDesde(t domain.Totales) totalesXML {
	out := totalesXML{MntTotal: t.MntTotal}
	if t.MntNeto > 0 {
		neto := t.MntNeto
		tasa := t.TasaIVA
		iva := t.IVA
		out.MntNeto = &neto
		out.TasaIVA = &tasa
		out.IVA = &iva
	}
	if t.MntExe > 0 {
		exe := t.MntExe
		out.MntExe = &exe
	}
	return out
}

// El orden de Transporte sigue la secuencia del Formato DTE 2.5 (febrero 2026):
// Patente, PatenteCarro, RUTTrans, Chofer, DirDest, CmnaDest, CiudadDest,
// FchSalida, HraSalida y FchLlegada.
type dteXML struct {
	XMLName   xml.Name     `xml:"DTE"`
	Version   string       `xml:"version,attr"`
	Documento documentoXML `xml:"Documento"`
}

type documentoXML struct {
	ID         string        `xml:"ID,attr"`
	Encabezado encabezadoXML `xml:"Encabezado"`
	Detalle    []detalleXML  `xml:"Detalle"`
}

type encabezadoXML struct {
	IdDoc      idDocXML      `xml:"IdDoc"`
	Emisor     emisorXML     `xml:"Emisor"`
	Receptor   receptorXML   `xml:"Receptor"`
	Transporte transporteXML `xml:"Transporte"`
	Totales    totalesXML    `xml:"Totales"`
}

type idDocXML struct {
	TipoDTE     int    `xml:"TipoDTE"`
	Folio       int64  `xml:"Folio"`
	FchEmis     string `xml:"FchEmis"`
	IndTraslado int    `xml:"IndTraslado"`
}

type emisorXML struct {
	RUTEmisor    string `xml:"RUTEmisor"`
	RznSoc       string `xml:"RznSoc"`
	GiroEmis     string `xml:"GiroEmis"`
	Acteco       string `xml:"Acteco"`
	DirOrigen    string `xml:"DirOrigen"`
	CmnaOrigen   string `xml:"CmnaOrigen"`
	CiudadOrigen string `xml:"CiudadOrigen,omitempty"`
}

type receptorXML struct {
	RUTRecep    string `xml:"RUTRecep"`
	RznSocRecep string `xml:"RznSocRecep"`
	GiroRecep   string `xml:"GiroRecep,omitempty"`
	DirRecep    string `xml:"DirRecep,omitempty"`
	CmnaRecep   string `xml:"CmnaRecep,omitempty"`
}

type transporteXML struct {
	Patente      string    `xml:"Patente"`
	PatenteCarro string    `xml:"PatenteCarro,omitempty"`
	RUTTrans     string    `xml:"RUTTrans,omitempty"`
	Chofer       choferXML `xml:"Chofer"`
	DirDest      string    `xml:"DirDest"`
	CmnaDest     string    `xml:"CmnaDest"`
	CiudadDest   string    `xml:"CiudadDest"`
	FchSalida    string    `xml:"FchSalida"`
	HraSalida    string    `xml:"HraSalida"`
	FchLlegada   string    `xml:"FchLlegada,omitempty"`
}

type choferXML struct {
	RUTChofer    string `xml:"RUTChofer"`
	NombreChofer string `xml:"NombreChofer"`
}

type totalesXML struct {
	MntNeto  *int64 `xml:"MntNeto,omitempty"`
	MntExe   *int64 `xml:"MntExe,omitempty"`
	TasaIVA  *int   `xml:"TasaIVA,omitempty"`
	IVA      *int64 `xml:"IVA,omitempty"`
	MntTotal int64  `xml:"MntTotal"`
}

type detalleXML struct {
	NroLinDet int    `xml:"NroLinDet"`
	IndExe    int    `xml:"IndExe,omitempty"`
	NmbItem   string `xml:"NmbItem"`
	QtyItem   string `xml:"QtyItem"`
	UnmdItem  string `xml:"UnmdItem"`
	PrcItem   string `xml:"PrcItem,omitempty"`
	MontoItem int64  `xml:"MontoItem"`
}
