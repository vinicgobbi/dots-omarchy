package tui

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// ponte implementa ui.UI para os módulos, que rodam numa goroutine à parte:
// cada mensagem vira um tea.Msg para a tela, e as perguntas esperam a
// resposta do usuário por um canal.
type ponte struct {
	p     *tea.Program
	log   *logs.Log
	saida *escritorLinhas
}

func novaPonte(log *logs.Log) *ponte {
	b := &ponte{log: log}
	b.saida = &escritorLinhas{ponte: b}
	return b
}

func (b *ponte) enviar(msg tea.Msg) { b.p.Send(msg) }

func (b *ponte) mensagem(tipo tipoLinha, nivel, prefixo, msg string) {
	b.log.Linha(nivel, msg)
	fmt.Fprintf(b.log, "%s %s\n", prefixo, msg)
	b.enviar(linhaMsg{tipo, msg})
}

func (b *ponte) Info(msg string)    { b.mensagem(linhaInfo, "INFO", "[*]", msg) }
func (b *ponte) Sucesso(msg string) { b.mensagem(linhaSucesso, "OK", "[+]", msg) }
func (b *ponte) Aviso(msg string)   { b.mensagem(linhaAviso, "AVISO", "[!]", msg) }

func (b *ponte) DryRun(texto string) {
	fmt.Fprintf(b.log, "[dry-run] %s\n", texto)
	b.enviar(linhaMsg{linhaDryRun, texto})
}

func (b *ponte) Saida() io.Writer { return b.saida }
func (b *ponte) Interativo() bool { return true }
func (b *ponte) Isolado() bool    { return true }

func (b *ponte) Confirmar(c ui.Confirmacao) bool {
	resposta := make(chan bool, 1)
	b.enviar(confirmarMsg{c, resposta})
	sim := <-resposta
	b.log.Linha("PERGUNTA", fmt.Sprintf("%s -> %v", c.Pergunta, sim))
	return sim
}

func (b *ponte) Exec(cmd *exec.Cmd) error {
	fim := make(chan error, 1)
	b.enviar(execMsg{cmd, fim})
	return <-fim
}

// escritorLinhas recebe a saída dos comandos, grava no log bruto e manda cada
// linha completa para a tela.
type escritorLinhas struct {
	mu       sync.Mutex
	pendente []byte
	ponte    *ponte
}

func (e *escritorLinhas) Write(p []byte) (int, error) {
	e.ponte.log.Write(p)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pendente = append(e.pendente, p...)
	for {
		i := bytes.IndexByte(e.pendente, '\n')
		if i < 0 {
			break
		}
		e.emitir(e.pendente[:i])
		e.pendente = e.pendente[i+1:]
	}
	return len(p), nil
}

// Flush manda o que sobrou sem quebra de linha no fim de um comando.
func (e *escritorLinhas) Flush() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.pendente) > 0 {
		e.emitir(e.pendente)
	}
	e.pendente = nil
}

func (e *escritorLinhas) emitir(linha []byte) {
	e.ponte.enviar(linhaMsg{linhaSaida, limparLinha(string(linha))})
}

// limparLinha deixa a linha segura para desenhar: fica só o que veio depois
// do último \r (barras de progresso), sem cores nem caracteres de controle.
func limparLinha(texto string) string {
	texto = strings.TrimRight(texto, "\r")
	if i := strings.LastIndexByte(texto, '\r'); i >= 0 {
		texto = texto[i+1:]
	}
	texto = strings.ReplaceAll(ansi.Strip(texto), "\t", "    ")
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, texto)
}
