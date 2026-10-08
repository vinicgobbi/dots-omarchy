// Package ui define o que os módulos usam para falar com o usuário. Há duas
// implementações: a TUI (pacote tui) e o modo texto (Texto, aqui), usado sem
// terminal interativo ou com --sem-tui.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/logs"
)

type UI interface {
	Info(msg string)
	Sucesso(msg string)
	Aviso(msg string)
	// DryRun mostra um comando ou alteração que não foi executado.
	DryRun(texto string)
	// Saida recebe stdout/stderr dos comandos.
	Saida() io.Writer
	// Interativo diz se há alguém no terminal para responder perguntas.
	Interativo() bool
	Confirmar(c Confirmacao) bool
	// Isolado diz se os comandos não interativos devem rodar sem terminal de
	// controle (na TUI, para um prompt inesperado não bagunçar a tela).
	Isolado() bool
}

// Confirmacao é uma pergunta s/N com contexto.
type Confirmacao struct {
	Titulo   string
	Texto    []string
	Itens    []string
	Pergunta string
	Perigo   bool
}

const (
	Verde    = "\033[0;32m"
	Amarelo  = "\033[1;33m"
	Vermelho = "\033[0;31m"
	Azul     = "\033[0;34m"
	Cinza    = "\033[1;30m"
	Negrito  = "\033[1m"
	NC       = "\033[0m"
)

// Texto é a interface em texto corrido, como a versão em Python.
type Texto struct {
	Log        *logs.Log
	interativo bool
	entrada    *bufio.Reader
}

func NovoTexto(log *logs.Log, interativo bool) *Texto {
	return &Texto{Log: log, interativo: interativo, entrada: bufio.NewReader(os.Stdin)}
}

func (t *Texto) mensagem(cor, prefixo, nivel, msg string) {
	fmt.Printf("%s%s %s%s\n", cor, prefixo, msg, NC)
	t.Log.Linha(nivel, msg)
	fmt.Fprintf(t.Log, "%s %s\n", prefixo, msg)
}

func (t *Texto) Info(msg string)    { t.mensagem(Amarelo, "[*]", "INFO", msg) }
func (t *Texto) Sucesso(msg string) { t.mensagem(Verde, "[+]", "OK", msg) }
func (t *Texto) Aviso(msg string)   { t.mensagem(Azul, "[!]", "AVISO", msg) }
func (t *Texto) Erro(msg string)    { t.mensagem(Vermelho, "[-]", "ERRO", msg) }

func (t *Texto) DryRun(texto string) {
	fmt.Printf("%s[dry-run]%s %s\n", Cinza, NC, texto)
	fmt.Fprintf(t.Log, "[dry-run] %s\n", texto)
}

func (t *Texto) Passo(atual, total int, titulo string) {
	fmt.Printf("%s%s[%d/%d]%s %s\n", Negrito, Azul, atual, total, NC, titulo)
	t.Log.Linha("PASSO", fmt.Sprintf("[%d/%d] %s", atual, total, titulo))
	fmt.Fprintf(t.Log, "[%d/%d] %s\n", atual, total, titulo)
}

func (t *Texto) Saida() io.Writer { return io.MultiWriter(os.Stdout, t.Log) }
func (t *Texto) Interativo() bool { return t.interativo }
func (t *Texto) Isolado() bool    { return false }

func (t *Texto) Banner() {
	fmt.Printf("%s%s\n==============================================\n", Negrito, Verde)
	fmt.Printf("   POST-INSTALL SETUP\n==============================================\n%s\n", NC)
}

// Pergunta lê uma linha do terminal.
func (t *Texto) Pergunta(texto string) (string, error) {
	fmt.Printf("%s[?] %s%s", Amarelo, texto, NC)
	linha, err := t.entrada.ReadString('\n')
	if err != nil && linha == "" {
		return "", err
	}
	return strings.TrimSpace(linha), nil
}

// SimNao é a pergunta s/N genérica.
func (t *Texto) SimNao(pergunta string) bool {
	resposta, err := t.Pergunta(pergunta + " [s/N]: ")
	if err != nil {
		return false
	}
	switch resposta {
	case "s", "S", "y", "Y":
		return true
	}
	return false
}

func (t *Texto) Confirmar(c Confirmacao) bool {
	cor := Negrito
	if c.Perigo {
		cor = Negrito + Vermelho
	}
	fmt.Printf("\n%s%s%s\n", cor, c.Titulo, NC)
	for _, linha := range c.Texto {
		fmt.Println(linha)
	}
	for _, item := range c.Itens {
		fmt.Printf("  - %s\n", item)
	}
	fmt.Println()
	return t.SimNao(c.Pergunta)
}
