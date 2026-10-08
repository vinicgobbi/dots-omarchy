// Package logs grava os dois arquivos de log de cada execução, em logs/ ao
// lado do projeto:
//
//	<prefixo>_<timestamp>.log      — só as mensagens info/aviso/erro/sucesso/passo,
//	                                 com hora e nível, sem código de cor.
//	<prefixo>_<timestamp>.raw.log  — transcrição bruta de tudo a partir de
//	                                 IniciarBruto (inclui a saída de
//	                                 pacman/yay/curl/flatpak etc.).
package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Log struct {
	mu           sync.Mutex
	Caminho      string
	CaminhoBruto string
	limpo        *os.File
	bruto        *os.File
}

// Iniciar prepara só o log "limpo"; o bruto começa em IniciarBruto, quando a
// execução dos módulos começa.
func Iniciar(raiz, prefixo string) (*Log, error) {
	pasta := filepath.Join(raiz, "logs")
	if err := os.MkdirAll(pasta, 0o755); err != nil {
		return nil, err
	}
	base := filepath.Join(pasta, prefixo+"_"+time.Now().Format("2006-01-02_150405"))
	limpo, err := os.Create(base + ".log")
	if err != nil {
		return nil, err
	}
	return &Log{Caminho: base + ".log", CaminhoBruto: base + ".raw.log", limpo: limpo}, nil
}

// Linha grava uma mensagem com timestamp e nível no log limpo e, se já
// iniciado, também no bruto.
func (l *Log) Linha(nivel, mensagem string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.limpo, "%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), nivel, mensagem)
}

// IniciarBruto abre o log bruto. A partir daqui, Write grava nele.
func (l *Log) IniciarBruto() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.OpenFile(l.CaminhoBruto, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	l.bruto = f
	return nil
}

// BrutoIniciado diz se o log bruto já está sendo gravado.
func (l *Log) BrutoIniciado() bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.bruto != nil
}

// Write grava no log bruto (descarta enquanto ele não foi iniciado).
func (l *Log) Write(p []byte) (int, error) {
	if l == nil {
		return len(p), nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.bruto != nil {
		l.bruto.Write(p)
	}
	return len(p), nil
}

func (l *Log) Fechar() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limpo.Close()
	if l.bruto != nil {
		l.bruto.Close()
	}
}
