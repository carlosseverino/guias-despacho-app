package validator

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"guias-despacho/internal/domain"

	v10 "github.com/go-playground/validator/v10"
)

const (
	fechaMinEmision  = "2002-08-01"
	fechaMinTraslado = "2003-04-01"
	fechaMax         = "2050-12-31"
)

var (
	reHora      = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d$`)
	reHoraCorta = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)
	reUnmd      = regexp.MustCompile(`^[A-Z0-9]{1,4}$`)
	validate    = nuevoValidador()
)

// Incumplimiento es un incumplimiento puntual de la Res. Ex. N° 154 o del formato DTE.
type Incumplimiento struct {
	Campo   string
	Mensaje string
}

// Etiqueta es el nombre legible del campo para el tablero de prevalidación.
func (i Incumplimiento) Etiqueta() string {
	return Etiqueta(i.Campo)
}

// Etiqueta traduce la ruta del campo a una etiqueta de formulario.
func Etiqueta(campo string) string {
	if strings.HasPrefix(campo, "Items[") {
		rest := strings.TrimPrefix(campo, "Items[")
		nStr, sub, _ := strings.Cut(rest, "].")
		n, _ := strconv.Atoi(nStr)
		nombre := map[string]string{
			"NmbItem":   "nombre",
			"QtyItem":   "cantidad",
			"UnmdItem":  "unidad de medida",
			"PrcItem":   "precio",
			"MontoItem": "monto",
			"IndExe":    "indicador exento o no facturable",
		}[sub]
		if nombre == "" {
			nombre = sub
		}
		return fmt.Sprintf("Ítem %d · %s", n+1, nombre)
	}
	if etiqueta, ok := etiquetas[campo]; ok {
		return etiqueta
	}
	return campo
}

var etiquetas = map[string]string{
	"Folio":          "Folio",
	"FchEmis":        "Fecha de emisión",
	"IndTraslado":    "Tipo de traslado",
	"RUTEmisor":      "RUT emisor",
	"RznSoc":         "Razón social emisor",
	"GiroEmis":       "Giro del emisor",
	"Acteco":         "Actividad económica",
	"RUTRecep":       "RUT receptor",
	"RznSocRecep":    "Razón social receptor",
	"GiroRecep":      "Giro del receptor",
	"DirRecep":       "Dirección del receptor",
	"CmnaRecep":      "Comuna del receptor",
	"RUTChofer":      "RUT del chofer",
	"NombreChofer":   "Nombre del chofer",
	"Patente":        "Patente del vehículo",
	"PatenteCarro":   "Patente del carro o remolque",
	"RUTTrans":       "RUT del transportista",
	"FchSalida":      "Fecha de salida",
	"HraSalida":      "Hora de salida",
	"FchLlegada":     "Fecha de llegada",
	"MotivoMultidia": "Motivo del traslado de más de un día",
	"DirOrigen":      "Dirección de origen",
	"CmnaOrigen":     "Comuna de origen",
	"CiudadOrigen":   "Ciudad de origen",
	"DirDest":        "Dirección de destino",
	"CmnaDest":       "Comuna de destino",
	"CiudadDest":     "Ciudad de destino",
	"Items":          "Detalle de mercadería",
}

var ordenCampos = []string{
	"Folio", "FchEmis", "IndTraslado",
	"RUTEmisor", "RznSoc", "GiroEmis", "Acteco",
	"RUTRecep", "RznSocRecep", "GiroRecep", "DirRecep", "CmnaRecep",
	"RUTChofer", "NombreChofer",
	"Patente", "PatenteCarro", "RUTTrans",
	"FchSalida", "HraSalida", "FchLlegada", "MotivoMultidia",
	"DirOrigen", "CmnaOrigen", "CiudadOrigen", "DirDest", "CmnaDest", "CiudadDest",
	"Items",
}

func nuevoValidador() *v10.Validate {
	v := v10.New()
	_ = v.RegisterValidation("rut_cl", func(fl v10.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || RutValido(s)
	})
	_ = v.RegisterValidation("patente_cl", func(fl v10.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || PatenteValida(s)
	})
	_ = v.RegisterValidation("hra_sii", func(fl v10.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || reHora.MatchString(s)
	})
	_ = v.RegisterValidation("unmd_cl", func(fl v10.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || reUnmd.MatchString(s)
	})
	return v
}

// NormalizarHora acepta HH:MM y lo completa a HH:MM:00.
func NormalizarHora(s string) string {
	s = strings.TrimSpace(s)
	if reHoraCorta.MatchString(s) {
		return s + ":00"
	}
	return s
}

// NormalizarGuia limpia RUT, patentes, hora, textos y calcula montos omitidos.
func NormalizarGuia(g *domain.Guia) {
	g.FchEmis = strings.TrimSpace(g.FchEmis)
	g.RUTEmisor = NormalizarRUT(g.RUTEmisor)
	g.RznSoc = strings.TrimSpace(g.RznSoc)
	g.GiroEmis = strings.TrimSpace(g.GiroEmis)
	g.Acteco = strings.TrimSpace(g.Acteco)

	g.RUTRecep = NormalizarRUT(g.RUTRecep)
	g.RznSocRecep = strings.TrimSpace(g.RznSocRecep)
	g.GiroRecep = strings.TrimSpace(g.GiroRecep)
	g.DirRecep = strings.TrimSpace(g.DirRecep)
	g.CmnaRecep = strings.TrimSpace(g.CmnaRecep)

	g.RUTChofer = NormalizarRUT(g.RUTChofer)
	g.NombreChofer = strings.TrimSpace(g.NombreChofer)

	g.Patente = NormalizarPatente(g.Patente)
	g.PatenteCarro = NormalizarPatente(g.PatenteCarro)
	g.RUTTrans = NormalizarRUT(g.RUTTrans)

	g.FchSalida = strings.TrimSpace(g.FchSalida)
	g.HraSalida = NormalizarHora(g.HraSalida)
	g.FchLlegada = strings.TrimSpace(g.FchLlegada)
	g.MotivoMultidia = strings.TrimSpace(g.MotivoMultidia)

	g.DirOrigen = strings.TrimSpace(g.DirOrigen)
	g.CmnaOrigen = strings.TrimSpace(g.CmnaOrigen)
	g.CiudadOrigen = strings.TrimSpace(g.CiudadOrigen)
	g.DirDest = strings.TrimSpace(g.DirDest)
	g.CmnaDest = strings.TrimSpace(g.CmnaDest)
	g.CiudadDest = strings.TrimSpace(g.CiudadDest)

	for i := range g.Items {
		it := &g.Items[i]
		it.NmbItem = strings.TrimSpace(it.NmbItem)
		it.UnmdItem = strings.ToUpper(strings.TrimSpace(it.UnmdItem))
		if !it.MontoInformado && !it.MontoInvalido && it.PrcItem > 0 && it.QtyItem > 0 {
			it.MontoItem = int64(math.Round(it.QtyItem * it.PrcItem))
			it.MontoInformado = true
		}
	}
}

// ValidarGuia aplica el formato DTE y las reglas de traslado de la Res. Ex. N° 154.
func ValidarGuia(g domain.Guia) []Incumplimiento {
	NormalizarGuia(&g)

	var errs []Incumplimiento
	err := validate.Struct(guiaTagsDesde(g))
	if err != nil {
		if ves, ok := err.(v10.ValidationErrors); ok {
			for _, fe := range ves {
				errs = append(errs, Incumplimiento{
					Campo:   campo(fe),
					Mensaje: traducir(fe),
				})
			}
		} else {
			errs = append(errs, Incumplimiento{Campo: "Guia", Mensaje: "No fue posible validar el documento."})
		}
	}
	errs = append(errs, reglasResolucion154(g)...)
	return ordenar(errs)
}

// guiaTags aísla las etiquetas de validator/v10 del modelo de dominio.
type guiaTags struct {
	Folio       int64  `validate:"required,min=1,max=9999999999"`
	FchEmis     string `validate:"required,datetime=2006-01-02"`
	IndTraslado int    `validate:"required,min=1,max=9"`

	RUTEmisor string `validate:"required,rut_cl"`
	RznSoc    string `validate:"required,max=100"`
	GiroEmis  string `validate:"required,max=80"`
	Acteco    string `validate:"required,numeric,len=6"`

	RUTRecep    string `validate:"required,rut_cl"`
	RznSocRecep string `validate:"required,max=100"`
	GiroRecep   string `validate:"omitempty,max=80"`
	DirRecep    string `validate:"omitempty,max=70"`
	CmnaRecep   string `validate:"omitempty,max=20"`

	RUTChofer    string `validate:"required,rut_cl"`
	NombreChofer string `validate:"required,max=30"`

	Patente      string `validate:"required,patente_cl"`
	PatenteCarro string `validate:"omitempty,patente_cl"`
	RUTTrans     string `validate:"omitempty,rut_cl"`

	FchSalida      string `validate:"required,datetime=2006-01-02"`
	HraSalida      string `validate:"required,hra_sii"`
	FchLlegada     string `validate:"omitempty,datetime=2006-01-02"`
	MotivoMultidia string `validate:"omitempty,max=120"`

	DirOrigen    string `validate:"required,max=60"`
	CmnaOrigen   string `validate:"required,max=20"`
	CiudadOrigen string `validate:"omitempty,max=20"`
	DirDest      string `validate:"required,max=70"`
	CmnaDest     string `validate:"required,max=20"`
	CiudadDest   string `validate:"required,max=20"`

	Items []itemTags `validate:"required,min=1,max=60,dive"`
}

type itemTags struct {
	NmbItem   string  `validate:"required,max=80"`
	QtyItem   float64 `validate:"gt=0"`
	UnmdItem  string  `validate:"required,unmd_cl"`
	PrcItem   float64 `validate:"gte=0"`
	MontoItem int64   `validate:"gte=0"`
	IndExe    int     `validate:"oneof=0 1 2"`
}

func guiaTagsDesde(g domain.Guia) guiaTags {
	items := make([]itemTags, len(g.Items))
	for i, it := range g.Items {
		items[i] = itemTags{
			NmbItem:   it.NmbItem,
			QtyItem:   it.QtyItem,
			UnmdItem:  it.UnmdItem,
			PrcItem:   it.PrcItem,
			MontoItem: it.MontoItem,
			IndExe:    it.IndExe,
		}
	}
	return guiaTags{
		Folio: g.Folio, FchEmis: g.FchEmis, IndTraslado: g.IndTraslado,
		RUTEmisor: g.RUTEmisor, RznSoc: g.RznSoc, GiroEmis: g.GiroEmis, Acteco: g.Acteco,
		RUTRecep: g.RUTRecep, RznSocRecep: g.RznSocRecep, GiroRecep: g.GiroRecep,
		DirRecep: g.DirRecep, CmnaRecep: g.CmnaRecep,
		RUTChofer: g.RUTChofer, NombreChofer: g.NombreChofer,
		Patente: g.Patente, PatenteCarro: g.PatenteCarro, RUTTrans: g.RUTTrans,
		FchSalida: g.FchSalida, HraSalida: g.HraSalida, FchLlegada: g.FchLlegada,
		MotivoMultidia: g.MotivoMultidia,
		DirOrigen:      g.DirOrigen, CmnaOrigen: g.CmnaOrigen, CiudadOrigen: g.CiudadOrigen,
		DirDest: g.DirDest, CmnaDest: g.CmnaDest, CiudadDest: g.CiudadDest,
		Items: items,
	}
}

func reglasResolucion154(g domain.Guia) []Incumplimiento {
	var errs []Incumplimiento
	agregar := func(campo, mensaje string) {
		errs = append(errs, Incumplimiento{Campo: campo, Mensaje: mensaje})
	}

	if fecha, ok := parseFecha(g.FchEmis); ok {
		if !enRango(fecha, fechaMinEmision, fechaMax) {
			agregar("FchEmis", "La fecha de emisión debe estar entre 2002-08-01 y 2050-12-31.")
		}
	}
	salida, salidaOK := parseFecha(g.FchSalida)
	if salidaOK && !enRango(salida, fechaMinTraslado, fechaMax) {
		agregar("FchSalida", "La fecha de salida debe estar entre 2003-04-01 y 2050-12-31.")
	}
	emision, emisionOK := parseFecha(g.FchEmis)
	if salidaOK && emisionOK && salida.Before(emision) {
		agregar("FchSalida", "La fecha de salida no puede ser anterior a la emisión: la guía debe existir antes de iniciar el traslado.")
	}

	llegada, llegadaOK := parseFecha(g.FchLlegada)
	if g.FchLlegada != "" && llegadaOK && !enRango(llegada, fechaMinTraslado, fechaMax) {
		agregar("FchLlegada", "La fecha de llegada debe estar entre 2003-04-01 y 2050-12-31.")
	}
	if g.FchLlegada != "" && !llegadaOK {
		// El formato inválido ya lo informa validator/v10.
	} else if g.Multidia && g.FchLlegada == "" {
		agregar("FchLlegada", "En un traslado de más de un día, FchLlegada es obligatoria.")
	} else if salidaOK && llegadaOK && llegada.Before(salida) {
		agregar("FchLlegada", "La fecha de llegada no puede ser anterior a la fecha de salida.")
	} else if salidaOK && llegadaOK && llegada.After(salida) && !g.Multidia {
		agregar("FchLlegada", "La llegada es posterior a la salida. Marque el traslado como de más de un día e indique el motivo.")
	} else if g.Multidia && salidaOK && llegadaOK && !llegada.After(salida) {
		agregar("FchLlegada", "Si el traslado dura más de un día, la fecha de llegada debe ser posterior a la de salida.")
	}
	if g.Multidia && g.MotivoMultidia == "" {
		agregar("MotivoMultidia", "Indique el motivo del traslado cuando el trayecto supera un día.")
	}

	switch {
	case !g.TransportePropio && g.RUTTrans == "":
		agregar("RUTTrans", "El RUT del transportista es obligatorio cuando el transporte lo realiza un tercero distinto del emisor.")
	case !g.TransportePropio && RutValido(g.RUTTrans) && g.RUTTrans == g.RUTEmisor:
		agregar("RUTTrans", "El transportista tercero debe tener un RUT distinto al del emisor.")
	case g.TransportePropio && g.RUTTrans != "" && RutValido(g.RUTTrans) && g.RUTTrans != g.RUTEmisor:
		agregar("RUTTrans", "Marcó transporte propio, pero el RUT del transportista no coincide con el emisor. Seleccione transportista tercero.")
	}

	if g.PatenteCarro != "" && g.Patente != "" && g.PatenteCarro == g.Patente {
		agregar("PatenteCarro", "La patente del carro o remolque debe ser distinta de la patente del vehículo principal.")
	}

	if len(g.Items) == 0 {
		return errs
	}
	for i, it := range g.Items {
		campoMonto := fmt.Sprintf("Items[%d].MontoItem", i)
		if it.MontoInvalido {
			continue
		}
		if it.IndExe == 0 && it.MontoItem <= 0 {
			agregar(campoMonto, "Indique un monto mayor que cero, o marque el ítem como exento o no facturable.")
		}
		if it.PrcItem > 0 && it.MontoInformado && it.QtyItem > 0 {
			esperado := int64(math.Round(it.QtyItem * it.PrcItem))
			diferencia := esperado - it.MontoItem
			if diferencia < 0 {
				diferencia = -diferencia
			}
			if diferencia > 1 {
				agregar(campoMonto, "El monto no coincide con cantidad por precio unitario.")
			}
		}
	}
	return errs
}

func parseFecha(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s)
	return t, err == nil
}

func enRango(t time.Time, min, max string) bool {
	desde, _ := time.Parse("2006-01-02", min)
	hasta, _ := time.Parse("2006-01-02", max)
	return !t.Before(desde) && !t.After(hasta)
}

func campo(fe v10.FieldError) string {
	ns := fe.StructNamespace()
	ns = strings.TrimPrefix(ns, "guiaTags.")
	return ns
}

func traducir(fe v10.FieldError) string {
	switch fe.Tag() {
	case "required":
		if fe.Field() == "Items" {
			return "Debe incluir al menos un ítem de mercadería."
		}
		return "Campo obligatorio."
	case "rut_cl":
		return "RUT inválido. Revise el cuerpo y el dígito verificador (módulo 11)."
	case "patente_cl":
		return "Patente inválida. Use el formato antiguo AA1111 o el nuevo AAAA11."
	case "hra_sii":
		return "La hora debe tener formato HH:MM:SS, entre 00:00:00 y 23:59:59."
	case "unmd_cl":
		return "Use una unidad de 1 a 4 caracteres alfanuméricos, por ejemplo UN, KG o LT."
	case "datetime":
		return "Fecha inválida. Use el formato AAAA-MM-DD."
	case "len", "numeric":
		if fe.Field() == "Acteco" {
			return "El código de actividad económica debe tener 6 dígitos."
		}
		return "El largo del campo no es válido."
	case "max":
		if fe.Kind().String() == "slice" || fe.Field() == "Items" {
			return "El detalle admite como máximo 60 ítems."
		}
		return fmt.Sprintf("Supera el largo máximo de %s caracteres.", fe.Param())
	case "min":
		if fe.Field() == "Folio" {
			return "El folio debe ser mayor que cero."
		}
		if fe.Field() == "Items" {
			return "Debe incluir al menos un ítem de mercadería."
		}
		if fe.Field() == "IndTraslado" {
			return "El tipo de traslado debe estar entre 1 y 9."
		}
		return "El valor es menor que el mínimo permitido."
	case "gt":
		return "La cantidad debe ser mayor que cero."
	case "gte":
		return "El valor no puede ser negativo."
	case "oneof":
		if fe.Field() == "IndExe" {
			return "El indicador debe ser afecto, exento o no facturable."
		}
		return "El valor no está permitido."
	default:
		return "El valor no cumple el formato exigido."
	}
}

// Combinar antepone los errores de lectura y oculta un duplicado de validator
// cuando el mismo campo ya trae un mensaje de parseo.
func Combinar(parseErrs, valErrs []Incumplimiento) []Incumplimiento {
	bloqueado := map[string]bool{}
	out := make([]Incumplimiento, 0, len(parseErrs)+len(valErrs))
	for _, e := range parseErrs {
		bloqueado[e.Campo] = true
		out = append(out, e)
	}
	for _, e := range valErrs {
		if bloqueado[e.Campo] {
			continue
		}
		out = append(out, e)
	}
	return ordenar(out)
}

func ordenar(errs []Incumplimiento) []Incumplimiento {
	rank := map[string]int{}
	for i, c := range ordenCampos {
		rank[c] = (i + 1) * 100
	}
	sort.SliceStable(errs, func(i, j int) bool {
		return peso(errs[i].Campo, rank) < peso(errs[j].Campo, rank)
	})
	return errs
}

func peso(campo string, rank map[string]int) int {
	if strings.HasPrefix(campo, "Items[") {
		rest := strings.TrimPrefix(campo, "Items[")
		nStr, sub, _ := strings.Cut(rest, "].")
		n, _ := strconv.Atoi(nStr)
		subOrden := map[string]int{
			"NmbItem": 1, "QtyItem": 2, "UnmdItem": 3, "PrcItem": 4, "MontoItem": 5, "IndExe": 6,
		}[sub]
		return rank["Items"] + n*10 + subOrden
	}
	if r, ok := rank[campo]; ok {
		return r
	}
	return 100000
}
