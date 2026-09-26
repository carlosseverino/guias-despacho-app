package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestInicio(t *testing.T) {
	app := nuevaApp(t)
	resp := pedir(t, app, httptest.NewRequest(http.MethodGet, "/", nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body := leer(t, resp)
	if !strings.Contains(body, "Guía de despacho") {
		t.Fatal("la portada no renderizó el título")
	}
}

func TestValidarRechazaRUT(t *testing.T) {
	app := nuevaApp(t)
	form := valoresEjemplo()
	form.Set("rut_chofer", "11.111.111-K")
	req := httptest.NewRequest(http.MethodPost, "/validar", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp := pedir(t, app, req)
	body := leer(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(body, "RUT del chofer") {
		t.Fatalf("el panel no mostró el RUT del chofer: %s", body)
	}
}

func TestDescargarXML(t *testing.T) {
	app := nuevaApp(t)
	req := httptest.NewRequest(http.MethodPost, "/descargar/xml", strings.NewReader(valoresEjemplo().Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := pedir(t, app, req)
	body := leer(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	if !strings.Contains(resp.Header.Get("Content-Disposition"), "guia-despacho-52-152.xml") {
		t.Fatalf("disposition: %s", resp.Header.Get("Content-Disposition"))
	}
	if !strings.Contains(body, "<TipoDTE>52</TipoDTE>") || !strings.Contains(body, "<HraSalida>08:30:00</HraSalida>") {
		t.Fatalf("xml inesperado: %s", body)
	}
}

func TestDescargarPDF(t *testing.T) {
	app := nuevaApp(t)
	req := httptest.NewRequest(http.MethodPost, "/descargar/pdf", strings.NewReader(valoresEjemplo().Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := pedir(t, app, req)
	body := leer(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", resp.StatusCode, body)
	}
	if !strings.HasPrefix(body, "%PDF") {
		t.Fatalf("no es un PDF: %.40q", body)
	}
}

func TestDescargaRechazadaConservaFormulario(t *testing.T) {
	app := nuevaApp(t)
	form := valoresEjemplo()
	form.Set("hra_salida", "")
	req := httptest.NewRequest(http.MethodPost, "/descargar/xml", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := pedir(t, app, req)
	body := leer(t, resp)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if !strings.Contains(body, "Hora de salida") || !strings.Contains(body, "Molinos del Valle SpA") {
		t.Fatalf("la respuesta no conservó el formulario ni el error: %s", body)
	}
}

func nuevaApp(t *testing.T) *fiber.App {
	t.Helper()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func pedir(t *testing.T, app *fiber.App, req *http.Request) *http.Response {
	t.Helper()
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func leer(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func valoresEjemplo() url.Values {
	v := url.Values{}
	v.Set("folio", "152")
	v.Set("fch_emis", "2026-11-02")
	v.Set("ind_traslado", "1")
	v.Set("rut_emisor", "76.123.456-0")
	v.Set("rzn_soc", "Molinos del Valle SpA")
	v.Set("giro_emis", "Elaboración de productos de molinería")
	v.Set("acteco", "106101")
	v.Set("rut_recep", "76.543.210-3")
	v.Set("rzn_soc_recep", "Panaderia Sur Ltda")
	v.Set("giro_recep", "Elaboración de pan")
	v.Set("dir_recep", "Av. Costanera 450")
	v.Set("cmna_recep", "Concepcion")
	v.Set("rut_chofer", "15.678.932-1")
	v.Set("nombre_chofer", "JUAN PEDRO SOTO LAGOS")
	v.Set("patente", "ABCD12")
	v.Set("patente_carro", "XY1234")
	v.Set("modo_transporte", "tercero")
	v.Set("rut_trans", "99.999.999-9")
	v.Set("fch_salida", "2026-11-02")
	v.Set("hra_salida", "08:30:00")
	v.Set("fch_llegada", "2026-11-03")
	v.Set("duracion", "multidia")
	v.Set("motivo_multidia", "Traslado interregional con pernocta en ruta")
	v.Set("dir_origen", "Camino Industrial 1200")
	v.Set("cmna_origen", "Pudahuel")
	v.Set("ciudad_origen", "Santiago")
	v.Set("dir_dest", "Av. Costanera 450")
	v.Set("cmna_dest", "Concepcion")
	v.Set("ciudad_dest", "Concepcion")
	v.Set("items[0].nombre", "Harina de trigo")
	v.Set("items[0].cantidad", "10")
	v.Set("items[0].unidad", "KG")
	v.Set("items[0].precio", "1200")
	v.Set("items[0].monto", "12000")
	v.Set("items[0].indexe", "0")
	v.Set("items[1].nombre", "Sacos de papel")
	v.Set("items[1].cantidad", "5")
	v.Set("items[1].unidad", "UN")
	v.Set("items[1].precio", "")
	v.Set("items[1].monto", "0")
	v.Set("items[1].indexe", "1")
	return v
}
