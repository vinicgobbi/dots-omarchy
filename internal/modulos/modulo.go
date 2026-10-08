// Package modulos registra as etapas do setup, na ordem de execução, e
// resolve as dependências entre elas.
package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

// Modulo é uma etapa do setup. Deps são ids de outros módulos que precisam
// rodar antes. Confirma marca módulos que perguntam antes de instalar algo
// (código de terceiros), para a revisão avisar.
type Modulo struct {
	ID        string
	Titulo    string
	Descricao string
	Executar  func(s *sistema.Sistema) error
	Deps      []string
	Confirma  bool
}

// Todos devolve os módulos na ordem de execução.
func Todos() []*Modulo {
	return []*Modulo{
		atualizacao, repositorios, pacotesBase, solaar, flatpaks, launchersJogos,
		tailscale, chromeGCM, bitwarden, phpExtensoes, ambienteUsuario, rustTools,
		claudeCode, ovpn, virtManager, vscodeNautilus, dotsOmarchy, webapps,
		limpeza, pluginsOmarchy, temasOmarchy,
	}
}

func indice(lista []*Modulo, id string) int {
	for i, m := range lista {
		if m.ID == id {
			return i
		}
	}
	return -1
}

// Puxado é uma dependência selecionada automaticamente.
type Puxado struct{ Por, Dep int }

// Marcar seleciona o módulo i e, em cadeia, as dependências dele (A depende
// de B, que depende de C). Devolve as dependências que estavam desmarcadas.
func Marcar(lista []*Modulo, sel []bool, i int) []Puxado {
	sel[i] = true
	var puxados []Puxado
	for _, dep := range lista[i].Deps {
		j := indice(lista, dep)
		if j >= 0 && !sel[j] {
			puxados = append(puxados, Puxado{i, j})
			puxados = append(puxados, Marcar(lista, sel, j)...)
		}
	}
	return puxados
}

// Resolver marca qualquer dependência de um módulo selecionado, mesmo que o
// usuário a tenha desmarcado. Sem isso, dá pra selecionar só "Extensões PHP
// (SQL Server)" sem "Instalar pacotes base" e a instalação falha no meio.
func Resolver(lista []*Modulo, sel []bool) []Puxado {
	var puxados []Puxado
	for i := range lista {
		if sel[i] {
			puxados = append(puxados, Marcar(lista, sel, i)...)
		}
	}
	return puxados
}

// Dependentes devolve os módulos selecionados que dependem de i.
func Dependentes(lista []*Modulo, sel []bool, i int) []int {
	var deps []int
	for j, m := range lista {
		if !sel[j] {
			continue
		}
		for _, d := range m.Deps {
			if d == lista[i].ID {
				deps = append(deps, j)
			}
		}
	}
	return deps
}

// Titulos traduz ids de dependência nos títulos dos módulos.
func Titulos(lista []*Modulo, ids []string) []string {
	var titulos []string
	for _, id := range ids {
		if j := indice(lista, id); j >= 0 {
			titulos = append(titulos, lista[j].Titulo)
		}
	}
	return titulos
}

// Selecionar transforma ids (do --modulos) numa seleção.
func Selecionar(lista []*Modulo, ids []string) ([]bool, error) {
	sel := make([]bool, len(lista))
	var desconhecidos []string
	for _, id := range ids {
		if j := indice(lista, id); j >= 0 {
			sel[j] = true
		} else {
			desconhecidos = append(desconhecidos, id)
		}
	}
	if len(desconhecidos) > 0 {
		return nil, fmt.Errorf("módulos desconhecidos: %v", desconhecidos)
	}
	return sel, nil
}
