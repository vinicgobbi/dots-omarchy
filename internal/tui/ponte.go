package tui

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// intervaloLote é o tempo máximo que uma linha espera para ir à tela. Uma saída
// rápida (pacman, yay) redesenha a tela no máximo uma vez nesse intervalo, em
// vez de uma vez por linha.
const intervaloLote = 50 * time.Millisecond

// ponte implementa ui.UI para os módulos, que rodam numa goroutine à parte:
// cada mensagem vira um tea.Msg para a tela, e as perguntas esperam a
// resposta do usuário por um canal.
type ponte struct {
	despachar func(tea.Msg)
	log       *logs.Log
	saida     *escritorLinhas

	// envio mantém a ordem: um lote nunca passa à frente da mensagem seguinte.
	envio    sync.Mutex
	mu       sync.Mutex
	lote     []linha
	agendado bool
}

func novaPonte(log *logs.Log) *ponte {
	b := &ponte{log: log}
	b.saida = &escritorLinhas{ponte: b}
	return b
}

// enviar manda uma mensagem à tela, depois das linhas que estiverem na fila.
func (b *ponte) enviar(msg tea.Msg) {
	b.envio.Lock()
	defer b.envio.Unlock()
	b.despacharLote()
	b.despachar(msg)
}

// enfileirar guarda a linha para o próximo lote.
func (b *ponte) enfileirar(l linha) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lote = append(b.lote, l)
	if !b.agendado {
		b.agendado = true
		time.AfterFunc(intervaloLote, func() {
			b.envio.Lock()
			defer b.envio.Unlock()
			b.despacharLote()
		})
	}
}

// despacharLote exige b.envio travado.
func (b *ponte) despacharLote() {
	b.mu.Lock()
	lote := b.lote
	b.lote, b.agendado = nil, false
	b.mu.Unlock()
	if len(lote) > 0 {
		b.despachar(loteMsg(lote))
	}
}

func (b *ponte) mensagem(tipo tipoLinha, nivel, prefixo, msg string) {
	b.log.Linha(nivel, msg)
	fmt.Fprintf(b.log, "%s %s\n", prefixo, msg)
	b.enfileirar(linha{tipo, msg})
}

func (b *ponte) Info(msg string)    { b.mensagem(linhaInfo, "INFO", "[*]", msg) }
func (b *ponte) Sucesso(msg string) { b.mensagem(linhaSucesso, "OK", "[+]", msg) }
func (b *ponte) Aviso(msg string)   { b.mensagem(linhaAviso, "AVISO", "[!]", msg) }

func (b *ponte) DryRun(texto string) {
	fmt.Fprintf(b.log, "[dry-run] %s\n", texto)
	b.enfileirar(linha{linhaDryRun, texto})
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

func (e *escritorLinhas) emitir(texto []byte) {
	e.ponte.enfileirar(linha{linhaSaida, limparLinha(string(texto))})
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
