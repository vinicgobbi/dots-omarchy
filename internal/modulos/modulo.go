// Package modulos registra as etapas do setup, na ordem de execução, e
// resolve as dependências entre elas. Cada módulo pertence a um grupo e é
// autocontido: instala o que usa, então pode ser marcado ou desmarcado sozinho.
package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

// Grupos, na ordem em que aparecem na interface e rodam.
const (
	GrupoSistema         = "Sistema"
	GrupoTerminal        = "Terminal"
	GrupoAplicativos     = "Aplicativos"
	GrupoDesenvolvimento = "Desenvolvimento"
	GrupoRede            = "Rede e virtualização"
	GrupoOmarchy         = "Omarchy e aparência"
	GrupoFinalizacao     = "Finalização"
)

// Modulo é uma etapa do setup. Deps são ids de outros módulos que precisam
// rodar antes. Confirma marca módulos que perguntam antes de instalar algo
// (código de terceiros), para a revisão avisar.
type Modulo struct {
	ID        string
	Grupo     string
	Titulo    string
	Descricao string
	Executar  func(s *sistema.Sistema) error
	Deps      []string
	Confirma  bool
}

// Todos devolve os módulos na ordem de execução (agrupados por Grupo).
func Todos() []*Modulo {
	return []*Modulo{
		// Sistema
		atualizacao, repositorios,
		// Terminal. O shell vem antes dos módulos que mexem no ~/.zshrc (node):
		// o instalador do Oh My Zsh troca o arquivo.
		shell,
		// Aplicativos
		chrome, bitwarden, flatpaks, launchersJogos, solaar,
		// Desenvolvimento
		gitGCM, docker, vscode, node, phpSQLServer, rust, claudeCode,
		// Rede e virtualização
		tailscale, ovpn, virtManager,
		// Omarchy e aparência. Plugins e temas ficam por último: são código de
		// terceiros e perguntam antes.
		nautilus, dotsOmarchy, webapps, pluginsOmarchy, temasOmarchy,
		// Finalização
		limpeza,
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
// usuário a tenha desmarcado. Sem isso, dá pra selecionar só "PHP e SQL Server"
// sem "Configurar repositórios" e a instalação falha no meio.
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
