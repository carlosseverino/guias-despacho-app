package main

import (
	"log"
	"os"

	"guias-despacho/internal/web"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app, err := web.New()
	if err != nil {
		log.Fatal(err)
	}
	addr := "127.0.0.1:" + port
	log.Printf("Guías de despacho local en http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}
