# Guías de despacho electrónicas (DTE 52)

Aplicación local en Go para capturar, validar y descargar un borrador de Guía de Despacho Electrónica (tipo 52), con las reglas de traslado de la Resolución Exenta N° 154 del SII. Esa resolución rige desde el **1 de noviembre de 2026**.

El XML sigue la estructura del Anexo Técnico DTE 2.5 (el atributo `version` del DTE se mantiene en `1.0`). El PDF417 del PDF es una **maqueta de referencia**. Este programa no timbra, no firma y no envía el documento al SII: para emitirlo de verdad se necesita certificado digital, código de autorización de folios y el canal de envío del Servicio.

## Ejecutar

Requiere Go 1.26 o superior (es la versión que declara `go.mod` por las dependencias actuales).

```bash
go run cmd/app/main.go
```

Abra [http://127.0.0.1:8080](http://127.0.0.1:8080). La aplicación escucha solo en localhost. La interfaz carga Tailwind, HTMX y Alpine desde una CDN, así que el navegador necesita salida a internet para verse con estilo.

Pruebas:

```bash
go test ./...
```

## Qué hace

- Formulario por secciones: emisor y receptor, conductor, transporte, fechas, origen y destino, detalle.
- Botón **Validar cumplimiento Res. 154**, con el detalle de cada incumplimiento.
- **Descargar XML (DTE)** y **Descargar PDF guía**. Si el documento no cumple, no se genera el archivo y el formulario conserva lo escrito.

Validaciones previas a la generación: RUT con módulo 11 (emisor, receptor, chofer y transportista), patente `AA1111` o `AAAA11`, chofer obligatorio, RUT de transportista cuando es un tercero distinto del emisor, hora de salida `HH:MM:SS`, fecha de llegada cuando el traslado dura más de un día, direcciones de origen y destino, y al menos un ítem con cantidad, unidad y monto o indicador de exento / no facturable.
