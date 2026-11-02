package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Primeira tela quando há uma release mais nova: atualizar agora (baixa, confere
// e substitui o pacote; depois é preciso executar de novo) ou ignorar e seguir
// com a versão atual.

type atualizadoMsg struct{ err error }

func (m modelo) teclaAtualizacao(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.atualizado {
		switch k.String() {
		case "q", "esc", "enter":
			return m, tea.Quit
		}
		return m, nil
	}
	switch k.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		m.cursorAtualizacao = 0
	case "down", "j":
		m.cursorAtualizacao = 1
	case "enter", " ":
		if m.cursorAtualizacao == 1 {
			m.tela = m.depoisAtualizacao
			return m, nil
		}
		m.atualizando, m.erroAtualizacao = true, ""
		aplicar := m.cfg.AplicarAtualizacao
		return m, func() tea.Msg { return atualizadoMsg{err: aplicar()} }
	}
	return m, nil
}

func (m modelo) viewAtualizacao() string {
	var b strings.Builder
	texto := lipgloss.NewStyle().Width(max(20, min(m.largura, 100)))

	if m.atualizado {
		b.WriteString(estiloDestaque.Render("✓ post-omarchy atualizado para "+m.cfg.VersaoNova) + "\n\n")
		b.WriteString(texto.Render("Esta execução ainda é a versão antiga. Execute de novo para usar a nova:") + "\n\n  ")
		b.WriteString(estiloNegrito.Render(m.cfg.ComandoReexecucao) + "\n")
		return b.String()
	}

	b.WriteString(estiloInfo.Bold(true).Render("Nova versão disponível: "+m.cfg.VersaoNova) +
		estiloSutil.Render(fmt.Sprintf("  (você está na %s)", m.cfg.Versao)) + "\n")
	b.WriteString(texto.Render(estiloSutil.Render("Atualizar baixa a release, confere o SHA-256 e substitui o binário " +
		"(e, no pacote da release, setup.sh, data/, dots/, assets/ e os READMEs — os .ovpn, os logs e o CLAUDE.md " +
		"ficam intactos). Depois é preciso executar de novo.")))
	b.WriteString("\n\n")

	if m.atualizando {
		b.WriteString(m.spinner.View() + " Baixando e conferindo " + m.cfg.VersaoNova + "…\n\n")
		b.WriteString(estiloSutil.Render("Aguarde: interromper agora deixaria o pacote pela metade.") + "\n")
		return b.String()
	}

	for i, opcao := range []string{"Atualizar agora", "Ignorar e continuar com a " + m.cfg.Versao} {
		if i == m.cursorAtualizacao {
			b.WriteString(estiloDestaque.Render("› "+opcao) + "\n")
		} else {
			b.WriteString("  " + opcao + "\n")
		}
	}
	if m.erroAtualizacao != "" {
		b.WriteString("\n" + texto.Render(estiloErro.Render("✗ Falha ao atualizar: "+m.erroAtualizacao)) + "\n")
	}
	return b.String()
}
