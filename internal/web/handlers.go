package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"

	"guias-despacho/internal/domain"
	"guias-despacho/internal/generator"
	"guias-despacho/internal/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

//go:embed templates/*.html
var templateFS embed.FS

// Resumen es el tablero que se muestra cuando la guía cumple la resolución.
type Resumen struct {
	Neto                string
	Exento              string
	IVA                 string
	Total               string
	NoFacturable        string
	MuestraExento       bool
	MuestraIVA          bool
	MuestraNoFacturable bool
	Items               int
	Traslado            string
	Ruta                string
	Salida              string
}

// Page es el modelo de la vista.
type Page struct {
	Form     Formulario
	Errores  []validator.Incumplimiento
	Validado bool
	Resumen  *Resumen
}

// Mal indica si el campo debe marcarse en el formulario.
func (p Page) Mal(campo string) bool {
	for _, e := range p.Errores {
		if e.Campo == campo {
			return true
		}
	}
	return false
}

// ItemsJSON embebe el detalle para Alpine sin escapar el objeto.
func (p Page) ItemsJSON() template.JS {
	items := p.Form.Items
	if items == nil {
		items = []ItemForm{}
	}
	b, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return template.JS(b)
}

// Tipos expone el catálogo de IndTraslado a la plantilla.
func (p Page) Tipos() []domain.TipoTraslado {
	return domain.TiposTraslado()
}

// New construye la aplicación HTTP local.
func New() (*fiber.App, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("plantillas: %w", err)
	}
	s := &server{tmpl: tmpl}
	app := fiber.New(fiber.Config{
		AppName:               "Guias de despacho local",
		DisableStartupMessage: true,
		BodyLimit:             2 * 1024 * 1024,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("error: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("No fue posible completar la operación.")
		},
	})
	app.Use(recover.New())
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "no-referrer")
		return c.Next()
	})
	app.Get("/", s.inicio)
	app.Post("/validar", s.validar)
	app.Post("/descargar/xml", s.descargarXML)
	app.Post("/descargar/pdf", s.descargarPDF)
	return app, nil
}

type server struct {
	tmpl *template.Template
}

func (s *server) inicio(c *fiber.Ctx) error {
	return s.render(c, "page", nuevaPagina(formularioVacio(), nil, false, domain.Guia{}))
}

func (s *server) validar(c *fiber.Ctx) error {
	f, g, parseErrs := leerFormulario(c)
	errs := validator.Combinar(parseErrs, validator.ValidarGuia(g))
	pagina := nuevaPagina(f, errs, true, g)
	if c.Get("HX-Request") == "true" {
		return s.render(c, "panel", pagina)
	}
	return s.render(c, "page", pagina)
}

func (s *server) descargarXML(c *fiber.Ctx) error {
	f, g, errs := evaluar(c)
	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return s.render(c, "page", nuevaPagina(f, errs, true, g))
	}
	xmlDoc, err := generator.GenerarXML(g)
	if err != nil {
		return err
	}
	c.Attachment(fmt.Sprintf("guia-despacho-52-%d.xml", g.Folio))
	c.Set(fiber.HeaderContentType, "application/xml; charset=utf-8")
	return c.Send(xmlDoc)
}

func (s *server) descargarPDF(c *fiber.Ctx) error {
	f, g, errs := evaluar(c)
	if len(errs) > 0 {
		c.Status(fiber.StatusUnprocessableEntity)
		return s.render(c, "page", nuevaPagina(f, errs, true, g))
	}
	pdf, err := generator.GenerarPDF(g)
	if err != nil {
		return err
	}
	c.Attachment(fmt.Sprintf("guia-despacho-52-%d.pdf", g.Folio))
	c.Set(fiber.HeaderContentType, "application/pdf")
	return c.Send(pdf)
}

func evaluar(c *fiber.Ctx) (Formulario, domain.Guia, []validator.Incumplimiento) {
	f, g, parseErrs := leerFormulario(c)
	return f, g, validator.Combinar(parseErrs, validator.ValidarGuia(g))
}

func nuevaPagina(f Formulario, errs []validator.Incumplimiento, validado bool, g domain.Guia) Page {
	p := Page{Form: f, Errores: errs, Validado: validado}
	if !validado || len(errs) > 0 {
		return p
	}
	validator.NormalizarGuia(&g)
	t := g.Totales()
	p.Resumen = &Resumen{
		Neto:                domain.FormatoPesos(t.MntNeto),
		Exento:              domain.FormatoPesos(t.MntExe),
		IVA:                 domain.FormatoPesos(t.IVA),
		Total:               domain.FormatoPesos(t.MntTotal),
		NoFacturable:        domain.FormatoPesos(t.MntNoFacturable),
		MuestraExento:       t.MntExe > 0,
		MuestraIVA:          t.MntNeto > 0,
		MuestraNoFacturable: t.MntNoFacturable > 0,
		Items:               len(g.Items),
		Traslado:            fmt.Sprintf("%d · %s", g.IndTraslado, domain.GlosaTraslado(g.IndTraslado)),
		Ruta:                g.CmnaOrigen + " → " + g.CmnaDest,
		Salida:              g.FchSalida + " " + g.HraSalida,
	}
	return p
}

func (s *server) render(c *fiber.Ctx, name string, data any) error {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
	c.Set("Cache-Control", "no-store")
	return c.Send(buf.Bytes())
}
