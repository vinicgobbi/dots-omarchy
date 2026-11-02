package main

import (
	"slices"
	"testing"

	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
)

func TestSelecaoInicial(t *testing.T) {
	lista := []*modulos.Modulo{{ID: "a"}, {ID: "b", Deps: []string{"a"}}, {ID: "c"}}

	sel, _, err := selecaoInicial(lista, "", "")
	if err != nil || sel != nil {
		t.Errorf("sem flags: sel = %v, err = %v", sel, err)
	}
	sel, _, err = selecaoInicial(lista, "c", "")
	if err != nil || !slices.Equal(sel, []bool{false, false, true}) {
		t.Errorf("--modulos: sel = %v, err = %v", sel, err)
	}
	sel, notas, err := selecaoInicial(lista, "", "a")
	if err != nil || !slices.Equal(sel, []bool{false, false, true}) || len(notas) != 1 {
		t.Errorf("--pular: sel = %v, notas = %v, err = %v", sel, notas, err)
	}
	if _, _, err := selecaoInicial(lista, "x", ""); err == nil {
		t.Error("--modulos aceitou id desconhecido")
	}
	if _, _, err := selecaoInicial(lista, "", "x"); err == nil {
		t.Error("--pular aceitou id desconhecido")
	}
}

func TestResolverTema(t *testing.T) {
	temas := []string{"https://x/omarchy-um-theme", "https://x/omarchy-dois-theme"}
	casos := []struct {
		valor, url string
		escolhido  bool
		erro       bool
	}{
		{"", "", false, false},
		{"none", "", true, false},
		{"um", temas[0], true, false},
		{temas[1], temas[1], true, false},
		{"tres", "", false, true},
	}
	for _, c := range casos {
		url, escolhido, err := resolverTema(temas, c.valor)
		if url != c.url || escolhido != c.escolhido || (err != nil) != c.erro {
			t.Errorf("resolverTema(%q) = %q, %v, %v", c.valor, url, escolhido, err)
		}
	}
}
