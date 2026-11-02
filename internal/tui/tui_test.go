package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
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

func teclaTipo(m modelo, k tea.KeyType) (modelo, tea.Cmd) {
	novo, cmd := m.Update(tea.KeyMsg{Type: k})
	return novo.(modelo), cmd
}

func letra(m modelo, r rune) modelo {
	novo, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	return novo.(modelo)
}

func TestLinhasSelecao(t *testing.T) {
	lista := []*modulos.Modulo{{ID: "a", Grupo: "G1"}, {ID: "b", Grupo: "G1"}, {ID: "c", Grupo: "G2"}}
	esperado := []linhaSelecao{{mod: -1, grupo: "G1"}, {mod: 0}, {mod: 1}, {mod: -1, grupo: "G2"}, {mod: 2}}
	if obtido := linhasSelecao(lista); !reflect.DeepEqual(obtido, esperado) {
		t.Errorf("linhasSelecao = %v, esperado %v", obtido, esperado)
	}
}

func TestTelaTema(t *testing.T) {
	cfg := configTeste(t)
	m := novoModelo(cfg, nil)
	if m.tela != telaTema {
		t.Fatalf("tela = %v, esperado a de tema", m.tela)
	}
	if m.cursorTema != 1 {
		t.Errorf("o cursor deveria começar no último tema, veio %d", m.cursorTema)
	}

	// Escolher um tema segue para os módulos, com ele ativo.
	m, _ = teclaTipo(m, tea.KeyUp)
	m, _ = teclaTipo(m, tea.KeyEnter)
	if m.tela != telaModulos || m.cfg.TemaAtivo != cfg.Dados.Themes[0] {
		t.Errorf("tela = %v, tema = %q", m.tela, m.cfg.TemaAtivo)
	}

	// esc nos módulos volta à pergunta do tema.
	m, _ = teclaTipo(m, tea.KeyEsc)
	if m.tela != telaTema {
		t.Errorf("esc nos módulos: tela = %v", m.tela)
	}

	// "Não instalar" desmarca o módulo.
	m = novoModelo(cfg, nil)
	m, _ = teclaTipo(m, tea.KeyDown)
	m, _ = teclaTipo(m, tea.KeyEnter)
	if m.marcado("temas_omarchy") || m.cfg.TemaAtivo != "" || m.tela != telaModulos {
		t.Errorf("marcado = %v, tema = %q, tela = %v", m.marcado("temas_omarchy"), m.cfg.TemaAtivo, m.tela)
	}

	// Com --tema, não pergunta.
	cfg.TemaEscolhido = true
	if m = novoModelo(cfg, nil); m.tela != telaModulos {
		t.Errorf("com --tema: tela = %v", m.tela)
	}
}

func TestDesmarcarDependenciaEmCascata(t *testing.T) {
	cfg := configTeste(t)
	cfg.TemaEscolhido = true
	m := novoModelo(cfg, nil)
	m = letra(m, ' ') // cursor em "Base", da qual "App" depende
	if m.marcado("base") || m.marcado("app") {
		t.Errorf("base = %v, app = %v", m.marcado("base"), m.marcado("app"))
	}
	if !strings.Contains(m.nota, "App") {
		t.Errorf("nota = %q", m.nota)
	}
}

func TestAtalhosDaRevisao(t *testing.T) {
	cfg := configTeste(t)
	cfg.TemaEscolhido = true
	m := novoModelo(cfg, nil)
	m, _ = teclaTipo(m, tea.KeyEnter)
	if m.tela != telaRevisao {
		t.Fatalf("tela = %v", m.tela)
	}
	if m = letra(m, 'n'); m.tela != telaModulos {
		t.Errorf("n na revisão: tela = %v", m.tela)
	}
}

func modeloExecutando(t *testing.T) modelo {
	t.Helper()
	cfg := configTeste(t)
	cfg.TemaEscolhido = true
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

func TestAbortarPodeSerDesfeito(t *testing.T) {
	m := modeloExecutando(t)
	m, _ = teclaTipo(m, tea.KeyCtrlC)
	if !m.pedirAbortar {
		t.Fatal("o primeiro ctrl+c não pediu confirmação")
	}
	if !strings.Contains(m.View(), "ctrl+c de novo") {
		t.Error("o pedido de confirmação não aparece na tela")
	}
	// Rolar a saída não desiste do abort; esc desiste.
	m, _ = teclaTipo(m, tea.KeyUp)
	if !m.pedirAbortar {
		t.Error("rolar a saída desistiu do abort")
	}
	m, _ = teclaTipo(m, tea.KeyEsc)
	if m.pedirAbortar {
		t.Error("esc não desistiu do abort")
	}
}

func TestCtrlCNaConfirmacaoAparece(t *testing.T) {
	m := modeloExecutando(t)
	resposta := make(chan bool, 1)
	novo, _ := m.Update(confirmarMsg{ui.Confirmacao{Titulo: "Temas", Pergunta: "Instalar?"}, resposta})
	m = novo.(modelo)
	m, _ = teclaTipo(m, tea.KeyCtrlC)
	if !strings.Contains(m.View(), "ctrl+c de novo") {
		t.Error("o primeiro ctrl+c não mostra aviso na janela de confirmação")
	}
	m, _ = teclaTipo(m, tea.KeyCtrlC)
	if !m.abortado {
		t.Error("o segundo ctrl+c não abortou")
	}
	if <-resposta {
		t.Error("o abort deveria responder não à confirmação")
	}
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
		for _, tl := range []tela{telaTema, telaModulos, telaRevisao} {
			m.tela = tl
			if altura := strings.Count(m.View(), "\n") + 1; altura > tamanho.a {
				t.Errorf("%dx%d, tela %v: %d linhas", tamanho.l, tamanho.a, tl, altura)
			}
		}
	}
}

func TestListaMostraOsGrupos(t *testing.T) {
	cfg := configTeste(t)
	cfg.TemaEscolhido = true
	novo, _ := novoModelo(cfg, nil).Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	tela := novo.(modelo).View()
	for _, esperado := range []string{"G1", "G2", "Base", "Temas"} {
		if !strings.Contains(tela, esperado) {
			t.Errorf("a lista não mostra %q:\n%s", esperado, tela)
		}
	}
}

// As linhas vão em lote, sem passar à frente da mensagem seguinte.
func TestPonteMantemAOrdem(t *testing.T) {
	log, err := logs.Iniciar(t.TempDir(), "teste")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Fechar()
	var recebidas []tea.Msg
	b := novaPonte(log)
	b.despachar = func(msg tea.Msg) { recebidas = append(recebidas, msg) }

	b.Info("um")
	b.Aviso("dois")
	b.enviar(fimMsg{})
	if len(recebidas) != 2 {
		t.Fatalf("recebidas = %v", recebidas)
	}
	lote, ok := recebidas[0].(loteMsg)
	if !ok || len(lote) != 2 || lote[0].texto != "um" || lote[1].texto != "dois" {
		t.Errorf("lote = %v", recebidas[0])
	}
	if _, ok := recebidas[1].(fimMsg); !ok {
		t.Errorf("segunda mensagem = %v", recebidas[1])
	}
}

func TestPonteDespachaSozinha(t *testing.T) {
	log, err := logs.Iniciar(t.TempDir(), "teste")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Fechar()
	chegou := make(chan tea.Msg, 1)
	b := novaPonte(log)
	b.despachar = func(msg tea.Msg) { chegou <- msg }
	b.Info("um")
	select {
	case <-chegou:
	case <-time.After(2 * time.Second):
		t.Fatal("o lote não foi despachado")
	}
}
