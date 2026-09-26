package web

import (
	"fmt"
	"strconv"
	"strings"

	"guias-despacho/internal/domain"
	"guias-despacho/internal/validator"

	"github.com/gofiber/fiber/v2"
)

// ItemForm conserva el texto ingresado para volver a pintar el formulario.
type ItemForm struct {
	Nombre   string `json:"nombre"`
	Cantidad string `json:"cantidad"`
	Unidad   string `json:"unidad"`
	Precio   string `json:"precio"`
	Monto    string `json:"monto"`
	IndExe   string `json:"indexe"`
}

// Formulario es el estado de la pantalla de emisión.
type Formulario struct {
	Folio          string
	FchEmis        string
	IndTraslado    string
	RUTEmisor      string
	RznSoc         string
	GiroEmis       string
	Acteco         string
	RUTRecep       string
	RznSocRecep    string
	GiroRecep      string
	DirRecep       string
	CmnaRecep      string
	RUTChofer      string
	NombreChofer   string
	Patente        string
	PatenteCarro   string
	ModoTransporte string
	RUTTrans       string
	FchSalida      string
	HraSalida      string
	FchLlegada     string
	Duracion       string
	MotivoMultidia string
	DirOrigen      string
	CmnaOrigen     string
	CiudadOrigen   string
	DirDest        string
	CmnaDest       string
	CiudadDest     string
	Items          []ItemForm
}

func formularioVacio() Formulario {
	return Formulario{
		IndTraslado:    "1",
		ModoTransporte: "propio",
		Duracion:       "mismo_dia",
		Items:          []ItemForm{{Unidad: "UN", IndExe: "0"}},
	}
}

func leerFormulario(c *fiber.Ctx) (Formulario, domain.Guia, []validator.Incumplimiento) {
	f := Formulario{
		Folio:          c.FormValue("folio"),
		FchEmis:        c.FormValue("fch_emis"),
		IndTraslado:    c.FormValue("ind_traslado"),
		RUTEmisor:      c.FormValue("rut_emisor"),
		RznSoc:         c.FormValue("rzn_soc"),
		GiroEmis:       c.FormValue("giro_emis"),
		Acteco:         c.FormValue("acteco"),
		RUTRecep:       c.FormValue("rut_recep"),
		RznSocRecep:    c.FormValue("rzn_soc_recep"),
		GiroRecep:      c.FormValue("giro_recep"),
		DirRecep:       c.FormValue("dir_recep"),
		CmnaRecep:      c.FormValue("cmna_recep"),
		RUTChofer:      c.FormValue("rut_chofer"),
		NombreChofer:   c.FormValue("nombre_chofer"),
		Patente:        c.FormValue("patente"),
		PatenteCarro:   c.FormValue("patente_carro"),
		ModoTransporte: c.FormValue("modo_transporte"),
		RUTTrans:       c.FormValue("rut_trans"),
		FchSalida:      c.FormValue("fch_salida"),
		HraSalida:      c.FormValue("hra_salida"),
		FchLlegada:     c.FormValue("fch_llegada"),
		Duracion:       c.FormValue("duracion"),
		MotivoMultidia: c.FormValue("motivo_multidia"),
		DirOrigen:      c.FormValue("dir_origen"),
		CmnaOrigen:     c.FormValue("cmna_origen"),
		CiudadOrigen:   c.FormValue("ciudad_origen"),
		DirDest:        c.FormValue("dir_dest"),
		CmnaDest:       c.FormValue("cmna_dest"),
		CiudadDest:     c.FormValue("ciudad_dest"),
		Items:          leerItems(c),
	}
	if f.ModoTransporte != "tercero" {
		f.ModoTransporte = "propio"
	}
	if f.Duracion != "multidia" {
		f.Duracion = "mismo_dia"
	}
	if f.IndTraslado == "" {
		f.IndTraslado = "1"
	}
	g, errs := aGuia(f)
	return f, g, errs
}

func leerItems(c *fiber.Ctx) []ItemForm {
	args := c.Request().PostArgs()
	var items []ItemForm
	for i := 0; i < 60; i++ {
		pref := fmt.Sprintf("items[%d].", i)
		presente := args.Has(pref+"nombre") || args.Has(pref+"cantidad") || args.Has(pref+"unidad") ||
			args.Has(pref+"precio") || args.Has(pref+"monto") || args.Has(pref+"indexe")
		if !presente {
			break
		}
		unidad := string(args.Peek(pref + "unidad"))
		if strings.TrimSpace(unidad) == "" {
			unidad = "UN"
		}
		indexe := string(args.Peek(pref + "indexe"))
		if indexe == "" {
			indexe = "0"
		}
		items = append(items, ItemForm{
			Nombre:   string(args.Peek(pref + "nombre")),
			Cantidad: string(args.Peek(pref + "cantidad")),
			Unidad:   unidad,
			Precio:   string(args.Peek(pref + "precio")),
			Monto:    string(args.Peek(pref + "monto")),
			IndExe:   indexe,
		})
	}

	var utiles []ItemForm
	for _, it := range items {
		if !filaVacia(it) {
			utiles = append(utiles, it)
		}
	}
	if len(utiles) == 0 {
		return []ItemForm{{Unidad: "UN", IndExe: "0"}}
	}
	return utiles
}

func filaVacia(it ItemForm) bool {
	return strings.TrimSpace(it.Nombre) == "" &&
		strings.TrimSpace(it.Cantidad) == "" &&
		strings.TrimSpace(it.Precio) == "" &&
		strings.TrimSpace(it.Monto) == "" &&
		(strings.TrimSpace(it.Unidad) == "" || strings.EqualFold(strings.TrimSpace(it.Unidad), "UN")) &&
		(it.IndExe == "" || it.IndExe == "0")
}

func aGuia(f Formulario) (domain.Guia, []validator.Incumplimiento) {
	var errs []validator.Incumplimiento
	g := domain.Guia{
		FchEmis:          strings.TrimSpace(f.FchEmis),
		RUTEmisor:        f.RUTEmisor,
		RznSoc:           f.RznSoc,
		GiroEmis:         f.GiroEmis,
		Acteco:           f.Acteco,
		RUTRecep:         f.RUTRecep,
		RznSocRecep:      f.RznSocRecep,
		GiroRecep:        f.GiroRecep,
		DirRecep:         f.DirRecep,
		CmnaRecep:        f.CmnaRecep,
		RUTChofer:        f.RUTChofer,
		NombreChofer:     f.NombreChofer,
		Patente:          f.Patente,
		PatenteCarro:     f.PatenteCarro,
		TransportePropio: f.ModoTransporte != "tercero",
		RUTTrans:         f.RUTTrans,
		FchSalida:        f.FchSalida,
		HraSalida:        f.HraSalida,
		FchLlegada:       f.FchLlegada,
		Multidia:         f.Duracion == "multidia",
		MotivoMultidia:   f.MotivoMultidia,
		DirOrigen:        f.DirOrigen,
		CmnaOrigen:       f.CmnaOrigen,
		CiudadOrigen:     f.CiudadOrigen,
		DirDest:          f.DirDest,
		CmnaDest:         f.CmnaDest,
		CiudadDest:       f.CiudadDest,
	}

	if strings.TrimSpace(f.Folio) == "" {
		errs = append(errs, validator.Incumplimiento{Campo: "Folio", Mensaje: "Campo obligatorio."})
	} else if n, err := strconv.ParseInt(strings.TrimSpace(f.Folio), 10, 64); err != nil || n < 1 {
		errs = append(errs, validator.Incumplimiento{Campo: "Folio", Mensaje: "El folio debe ser un número entero mayor que cero."})
	} else {
		g.Folio = n
	}

	if n, err := strconv.Atoi(strings.TrimSpace(f.IndTraslado)); err != nil {
		errs = append(errs, validator.Incumplimiento{Campo: "IndTraslado", Mensaje: "Seleccione el tipo de traslado."})
	} else {
		g.IndTraslado = n
	}

	for i, it := range f.Items {
		linea, lineErrs := aItem(i, it)
		g.Items = append(g.Items, linea)
		errs = append(errs, lineErrs...)
	}
	return g, errs
}

func aItem(i int, it ItemForm) (domain.Item, []validator.Incumplimiento) {
	var errs []validator.Incumplimiento
	item := domain.Item{
		NmbItem:  it.Nombre,
		UnmdItem: it.Unidad,
	}
	campo := func(sub string) string {
		return fmt.Sprintf("Items[%d].%s", i, sub)
	}

	if cant := strings.TrimSpace(it.Cantidad); cant != "" {
		q, err := parseDecimal(cant)
		if err != nil {
			errs = append(errs, validator.Incumplimiento{Campo: campo("QtyItem"), Mensaje: "La cantidad debe ser un número mayor que cero."})
		} else {
			item.QtyItem = q
		}
	}

	if precio := strings.TrimSpace(it.Precio); precio != "" {
		p, err := parseDecimal(precio)
		if err != nil || p < 0 {
			errs = append(errs, validator.Incumplimiento{Campo: campo("PrcItem"), Mensaje: "El precio unitario debe ser un número mayor o igual que cero."})
		} else {
			item.PrcItem = p
		}
	}

	if monto := strings.TrimSpace(it.Monto); monto != "" {
		item.MontoInformado = true
		m, err := parseMonto(monto)
		if err != nil || m < 0 {
			item.MontoInvalido = true
			errs = append(errs, validator.Incumplimiento{Campo: campo("MontoItem"), Mensaje: "El monto debe ser un entero en pesos, sin decimales."})
		} else {
			item.MontoItem = m
		}
	}

	switch strings.TrimSpace(it.IndExe) {
	case "", "0":
		item.IndExe = 0
	case "1":
		item.IndExe = 1
	case "2":
		item.IndExe = 2
	default:
		item.IndExe = -1
	}
	return item, errs
}

func parseDecimal(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else {
		s = strings.ReplaceAll(s, ",", ".")
	}
	return strconv.ParseFloat(s, 64)
}

func parseMonto(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimPrefix(s, "$")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(s, 10, 64)
}
