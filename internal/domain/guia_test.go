package domain

import "testing"

func TestTotalesRedondeaIVA(t *testing.T) {
	g := Guia{Items: []Item{
		{MontoItem: 10, IndExe: 0},
		{MontoItem: 500, IndExe: 1},
		{MontoItem: 80, IndExe: 2},
	}}
	tot := g.Totales()
	if tot.MntNeto != 10 || tot.IVA != 2 || tot.MntExe != 500 || tot.MntNoFacturable != 80 || tot.MntTotal != 512 {
		t.Fatalf("%+v", tot)
	}
}

func TestFormatoPesos(t *testing.T) {
	if got := FormatoPesos(14280); got != "$14.280" {
		t.Fatalf("formato = %s", got)
	}
}
