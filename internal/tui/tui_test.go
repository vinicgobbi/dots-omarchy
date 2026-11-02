package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

func configTeste(t *testing.T) Config {
	t.Helper()
	log, err := logs.Iniciar(t.TempDir(), "teste")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(log.Fechar)
	return Config{
		Versao: "v1.0.0", Raiz: t.TempDir(), Log: log, NomeSistema: "Omarchy",
		Dados: &dados.Dados{Themes: []string{"https://x/omarchy-um-theme", "https://x/omarchy-dois-theme"}},
		Modulos: []*modulos.Modulo{
			{ID: "base", Grupo: "G1", Titulo: "Base"},
			{ID: "app", Grupo: "G1", Titulo: "App", Deps: []string{"base"}},
			{ID: "temas_omarchy", Grupo: "G2", Titulo: "Temas"},
		},
		Conta: &sistema.Conta{Nome: "ana", Home: "/home/ana"},
	}
}

func modeloExecutando(t *testing.T) modelo {
	t.Helper()
	cfg := configTeste(t)
	m := novoModelo(cfg, nil)
	m.fila = cfg.Modulos
	m.resultados = make([]modulos.Resultado, len(m.fila))
	for i, mod := range m.fila {
		m.resultados[i].Modulo = mod
	}
	m.tela, m.inicio = telaExecucao, time.Now()
	m.cancelar = func() {}
	return m
}

func TestEstadosNoResumo(t *testing.T) {
	m := modeloExecutando(t)
	finais := []modulos.Resultado{
		{Modulo: m.fila[0], Estado: modulos.Falhou, Erro: errors.New("boom")},
		{Modulo: m.fila[1], Estado: modulos.Pulado, Erro: errors.New("dependência 'base' não concluída")},
		{Modulo: m.fila[2], Estado: modulos.ConcluidoComAvisos, Avisos: []string{"cuidado"}},
	}
	for i, r := range finais {
		novo, _ := m.Update(moduloFimMsg{i, r})
		m = novo.(modelo)
	}
	novo, _ := m.Update(fimMsg{})
	m = novo.(modelo)
	if m.tela != telaResumo {
		t.Fatalf("tela = %v", m.tela)
	}
	tela := m.View()
	for _, esperado := range []string{"Concluído com falhas", "1 ok", "1 falha(s)", "1 pulado(s)", "boom", "cuidado", "1 aviso(s)"} {
		if !strings.Contains(tela, esperado) {
			t.Errorf("o resumo não mostra %q:\n%s", esperado, tela)
		}
	}
	terminal := Resumo(m.resultado(), false, "log", "bruto")
	for _, esperado := range []string{"Concluído com falhas", "boom", "cuidado", "log", "bruto"} {
		if !strings.Contains(terminal, esperado) {
			t.Errorf("o resumo do terminal não mostra %q:\n%s", esperado, terminal)
		}
	}
}

// Nenhuma tela pode quebrar nem passar da altura, mesmo em terminal pequeno.
func TestTelasCabemNoTerminal(t *testing.T) {
	cfg := configTeste(t)
	cfg.Modulos = modulos.Todos()
	for _, tamanho := range []struct{ l, a int }{{80, 24}, {120, 40}, {60, 20}} {
		m := novoModelo(cfg, nil)
		novo, _ := m.Update(tea.WindowSizeMsg{Width: tamanho.l, Height: tamanho.a})
		m = novo.(modelo)
		for _, tl := range []tela{telaModulos, telaRevisao} {
			m.tela = tl
			if altura := strings.Count(m.View(), "\n") + 1; altura > tamanho.a {
				t.Errorf("%dx%d, tela %v: %d linhas", tamanho.l, tamanho.a, tl, altura)
			}
		}
	}
}
