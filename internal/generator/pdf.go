package generator

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"

	"guias-despacho/internal/domain"
	"guias-despacho/internal/validator"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/pdf417"
	"github.com/jung-kurt/gofpdf"
)

// GenerarPDF construye la representación imprimible de trabajo.
// El PDF417 es una maqueta de referencia local: no es el timbre electrónico del SII.
func GenerarPDF(g domain.Guia) ([]byte, error) {
	validator.NormalizarGuia(&g)
	if errs := validator.ValidarGuia(g); len(errs) > 0 {
		return nil, DocumentoInvalidoError{Errores: errs}
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 12, 12)
	pdf.SetAutoPageBreak(true, 16)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	t := func(s string) string { return tr(s) }

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "", 7.5)
		pdf.SetTextColor(90, 84, 74)
		pdf.CellFormat(0, 5, t("Borrador local  ·  Guía de despacho electrónica tipo 52  ·  Sin timbre electrónico ni firma digital  ·  No válido ante el SII"), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()
	dibujarPortada(pdf, t, g)
	dibujarPartes(pdf, t, g)
	dibujarTraslado(pdf, t, g)
	dibujarDetalle(pdf, t, g)
	if err := dibujarReferencia(pdf, t, g); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("escribir PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func dibujarPortada(pdf *gofpdf.Fpdf, t func(string) string, g domain.Guia) {
	pdf.SetFillColor(20, 34, 28)
	pdf.Rect(0, 0, 210, 26, "F")
	pdf.SetTextColor(246, 241, 231)
	pdf.SetFont("Helvetica", "B", 15)
	pdf.SetXY(12, 7)
	pdf.Cell(130, 7, t("GUÍA DE DESPACHO ELECTRÓNICA"))
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetXY(12, 15)
	pdf.Cell(130, 5, t("DTE tipo 52   ·   Representación de trabajo para la ruta"))
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetXY(120, 9)
	pdf.CellFormat(78, 8, t(fmt.Sprintf("FOLIO  %d", g.Folio)), "", 0, "R", false, 0, "")

	pdf.SetFillColor(255, 244, 230)
	pdf.SetDrawColor(196, 98, 45)
	pdf.Rect(12, 30, 186, 14, "FD")
	pdf.SetTextColor(122, 52, 18)
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetXY(15, 32)
	pdf.MultiCell(180, 3.6, t("Prevalidación local según la Res. Ex. SII N°154, aplicable desde el 1 de noviembre de 2026. Este archivo no incluye código de autorización de folios, timbre electrónico ni firma digital, y no ha sido enviado al SII."), "", "L", false)
	pdf.SetY(48)
}

func dibujarPartes(pdf *gofpdf.Fpdf, t func(string) string, g domain.Guia) {
	y := pdf.GetY()
	emisor := [][2]string{
		{"RUT", g.RUTEmisor},
		{"Razón social", g.RznSoc},
		{"Giro", g.GiroEmis},
		{"Acteco", g.Acteco},
	}
	receptor := [][2]string{
		{"RUT", g.RUTRecep},
		{"Razón social", g.RznSocRecep},
		{"Giro", valorOGuion(g.GiroRecep)},
		{"Dirección", valorOGuion(g.DirRecep)},
		{"Comuna", valorOGuion(g.CmnaRecep)},
	}
	alto := columna(pdf, t, 12, y, 91, "Emisor", emisor)
	alto2 := columna(pdf, t, 107, y, 91, "Receptor", receptor)
	if alto2 > alto {
		alto = alto2
	}
	pdf.SetY(y + alto + 4)
}

func dibujarTraslado(pdf *gofpdf.Fpdf, t func(string) string, g domain.Guia) {
	y := pdf.GetY()
	llegada := "Mismo día, sin fecha de llegada distinta"
	if g.FchLlegada != "" {
		llegada = g.FchLlegada
	}
	transporte := [][2]string{
		{"Chofer", g.NombreChofer},
		{"RUT chofer", g.RUTChofer},
		{"Patente", g.Patente},
		{"Carro / remolque", valorOGuion(g.PatenteCarro)},
		{"Transportista", transportista(g)},
		{"Salida", g.FchSalida + "  " + g.HraSalida},
		{"Llegada", llegada},
		{"Tipo de traslado", fmt.Sprintf("%d  %s", g.IndTraslado, domain.GlosaTraslado(g.IndTraslado))},
	}
	if g.Multidia {
		transporte = append(transporte, [2]string{"Motivo", g.MotivoMultidia})
	}
	ruta := [][2]string{
		{"Origen", g.DirOrigen},
		{"Comuna origen", g.CmnaOrigen},
		{"Ciudad origen", valorOGuion(g.CiudadOrigen)},
		{"Destino", g.DirDest},
		{"Comuna destino", g.CmnaDest},
		{"Ciudad destino", g.CiudadDest},
		{"Emisión", g.FchEmis},
	}
	alto := columna(pdf, t, 12, y, 91, "Conductor y transporte", transporte)
	alto2 := columna(pdf, t, 107, y, 91, "Origen y destino", ruta)
	if alto2 > alto {
		alto = alto2
	}
	pdf.SetY(y + alto + 4)
}

func dibujarDetalle(pdf *gofpdf.Fpdf, t func(string) string, g domain.Guia) {
	pdf.SetFont("Helvetica", "B", 9)
	pdf.SetTextColor(20, 34, 28)
	pdf.Cell(0, 6, t("Detalle de mercadería"))
	pdf.Ln(7)

	anchos := []float64{8, 68, 16, 14, 26, 28, 26}
	cabeceras := []string{"#", "Descripción", "Cant.", "Un.", "P. unitario", "Monto", "Tratamiento"}
	pdf.SetFillColor(20, 34, 28)
	pdf.SetTextColor(246, 241, 231)
	pdf.SetFont("Helvetica", "B", 7.5)
	for i, c := range cabeceras {
		pdf.CellFormat(anchos[i], 6, t(c), "", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for i, it := range g.Items {
		if pdf.GetY() > 250 {
			pdf.AddPage()
		}
		if i%2 == 0 {
			pdf.SetFillColor(246, 241, 231)
		} else {
			pdf.SetFillColor(255, 252, 247)
		}
		pdf.SetTextColor(28, 25, 21)
		precio := "—"
		if it.PrcItem > 0 {
			precio = domain.FormatoPesos(int64(it.PrcItem))
		}
		vals := []string{
			fmt.Sprintf("%d", i+1),
			it.NmbItem,
			domain.FormatoCantidad(it.QtyItem),
			it.UnmdItem,
			precio,
			domain.FormatoPesos(it.MontoItem),
			domain.GlosaIndExe(it.IndExe),
		}
		align := []string{"C", "L", "R", "C", "R", "R", "L"}
		for c, v := range vals {
			pdf.CellFormat(anchos[c], 6.5, t(corta(v, anchos[c])), "", 0, align[c], true, 0, "")
		}
		pdf.Ln(-1)
	}

	tot := g.Totales()
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.SetTextColor(28, 25, 21)
	escribirTotal(pdf, t, "Monto neto", domain.FormatoPesos(tot.MntNeto))
	if tot.MntExe > 0 {
		escribirTotal(pdf, t, "Monto exento", domain.FormatoPesos(tot.MntExe))
	}
	if tot.MntNeto > 0 {
		escribirTotal(pdf, t, "IVA 19%", domain.FormatoPesos(tot.IVA))
	}
	if tot.MntNoFacturable > 0 {
		escribirTotal(pdf, t, "No facturable (fuera del total)", domain.FormatoPesos(tot.MntNoFacturable))
	}
	pdf.SetFont("Helvetica", "B", 10)
	escribirTotal(pdf, t, "Total", domain.FormatoPesos(tot.MntTotal))
	pdf.Ln(3)
}

func escribirTotal(pdf *gofpdf.Fpdf, t func(string) string, etiqueta, valor string) {
	pdf.SetX(112)
	pdf.CellFormat(50, 5.5, t(etiqueta), "", 0, "L", false, 0, "")
	pdf.CellFormat(36, 5.5, t(valor), "", 1, "R", false, 0, "")
}

func dibujarReferencia(pdf *gofpdf.Fpdf, t func(string) string, g domain.Guia) error {
	if pdf.GetY() > 230 {
		pdf.AddPage()
	}
	payload := fmt.Sprintf("BORRADOR-LOCAL|T52|F%d|E%s|%s|%s|%s", g.Folio, g.RUTEmisor, g.FchEmis, g.FchSalida, g.Patente)
	pngBytes, err := codigoPDF417(payload)
	if err != nil {
		return err
	}
	opt := gofpdf.ImageOptions{ImageType: "PNG"}
	pdf.RegisterImageOptionsReader("pdf417", opt, bytes.NewReader(pngBytes))

	pdf.SetY(pdf.GetY() + 2)
	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(20, 34, 28)
	pdf.Cell(0, 5, t("Código PDF417 de referencia local"))
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "", 7.5)
	pdf.SetTextColor(70, 64, 54)
	pdf.MultiCell(186, 3.6, t("Maqueta visual para la representación de trabajo. No corresponde al timbre electrónico (TED) del SII y no permite acreditar el documento en ruta."), "", "L", false)
	pdf.Ln(1)
	y := pdf.GetY()
	pdf.ImageOptions("pdf417", 12, y, 92, 0, false, opt, 0, "")
	if cfg, err := png.DecodeConfig(bytes.NewReader(pngBytes)); err == nil && cfg.Width > 0 {
		pdf.SetY(y + 92*float64(cfg.Height)/float64(cfg.Width) + 2)
	}
	return nil
}

func codigoPDF417(payload string) ([]byte, error) {
	codigo, err := pdf417.Encode(payload, 2)
	if err != nil {
		return nil, fmt.Errorf("generar PDF417: %w", err)
	}
	limites := codigo.Bounds()
	escalado, err := barcode.Scale(codigo, limites.Dx()*3, limites.Dy()*3)
	if err != nil {
		return nil, fmt.Errorf("escalar PDF417: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, imagen8(escalado)); err != nil {
		return nil, fmt.Errorf("codificar PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// imagen8 pasa el código a PNG gris de 8 bits. gofpdf no abre PNG de 16 bits
// ni siempre acepta un canal alfa, y el escalado del PDF417 puede producirlos.
func imagen8(src image.Image) *image.Gray {
	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func columna(pdf *gofpdf.Fpdf, t func(string) string, x, y, w float64, titulo string, filas [][2]string) float64 {
	alto := 8 + float64(len(filas))*5.2 + 2
	pdf.SetFillColor(255, 252, 247)
	pdf.SetDrawColor(212, 203, 186)
	pdf.Rect(x, y, w, alto, "FD")
	pdf.SetXY(x+3, y+2)
	pdf.SetFont("Helvetica", "B", 8.5)
	pdf.SetTextColor(20, 34, 28)
	pdf.Cell(w-6, 5, t(titulo))
	yy := y + 8
	for _, fila := range filas {
		pdf.SetXY(x+3, yy)
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(110, 102, 90)
		pdf.Cell(28, 4.5, t(fila[0]))
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(28, 25, 21)
		pdf.Cell(w-34, 4.5, t(corta(fila[1], w-34)))
		yy += 5.2
	}
	return alto
}

func transportista(g domain.Guia) string {
	if g.TransportePropio && g.RUTTrans == "" {
		return "Propio (" + g.RUTEmisor + ")"
	}
	if g.TransportePropio {
		return "Propio (" + g.RUTTrans + ")"
	}
	return g.RUTTrans
}

func valorOGuion(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func corta(s string, anchoMM float64) string {
	max := int(anchoMM / 1.7)
	if max < 8 {
		max = 8
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "..."
}
