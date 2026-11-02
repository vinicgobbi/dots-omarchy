package modulos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// Estado é a situação de um módulo na execução.
type Estado int

const (
	Pendente Estado = iota
	Rodando
	Concluido
	ConcluidoComAvisos
	Falhou
	Pulado
	Cancelado
)

// Resultado é o que aconteceu com um módulo.
type Resultado struct {
	Modulo  *Modulo
	Estado  Estado
	Erro    error
	Duracao time.Duration
	Avisos  []string
}

// Observador recebe o andamento da execução. Fim também é chamado para os
// módulos que não chegaram a rodar (pulados e cancelados).
type Observador struct {
	Inicio func(n int)
	Fim    func(n int, r Resultado)
}

// avisosUI repassa tudo para a UI real e guarda os avisos do módulo atual.
type avisosUI struct {
	ui.UI
	mu     sync.Mutex
	avisos []string
}

func (a *avisosUI) Aviso(msg string) {
	a.mu.Lock()
	a.avisos = append(a.avisos, msg)
	a.mu.Unlock()
	a.UI.Aviso(msg)
}

// Rodar executa os módulos em ordem. Se um falhar, os seguintes continuam,
// exceto os que dependem (direta ou indiretamente) dele, que são pulados; com
// falhaRapida, para tudo no primeiro erro.
func Rodar(ctx context.Context, s *sistema.Sistema, fila []*Modulo, falhaRapida bool, o Observador) []Resultado {
	resultados := make([]Resultado, len(fila))
	quebrados := map[string]bool{} // ids que falharam ou foram pulados
	parar := false

	for n, m := range fila {
		r := Resultado{Modulo: m}
		dep := primeiraQuebrada(m, quebrados)
		switch {
		case parar || ctx.Err() != nil:
			r.Estado = Cancelado
		case dep != "":
			r.Estado = Pulado
			r.Erro = fmt.Errorf("dependência '%s' não concluída", dep)
			quebrados[m.ID] = true
			s.UI.Aviso(fmt.Sprintf("Pulando '%s': %v.", m.Titulo, r.Erro))
		default:
			o.Inicio(n)
			r = rodarUm(ctx, s, m)
			switch r.Estado {
			case Falhou:
				quebrados[m.ID] = true
				parar = falhaRapida
			case Cancelado:
				parar = true
			}
		}
		resultados[n] = r
		o.Fim(n, r)
	}
	return resultados
}

func primeiraQuebrada(m *Modulo, quebrados map[string]bool) string {
	for _, d := range m.Deps {
		if quebrados[d] {
			return d
		}
	}
	return ""
}

func rodarUm(ctx context.Context, s *sistema.Sistema, m *Modulo) Resultado {
	real := s.UI
	avisos := &avisosUI{UI: real}
	s.UI = avisos
	defer func() { s.UI = real }()

	inicio := time.Now()
	err := rodarSeguro(s, m)
	r := Resultado{Modulo: m, Erro: err, Duracao: time.Since(inicio), Avisos: avisos.avisos}
	switch {
	case err != nil && (ctx.Err() != nil || errors.Is(err, context.Canceled)):
		r.Estado = Cancelado
	case err != nil:
		r.Estado = Falhou
	case len(r.Avisos) > 0:
		r.Estado = ConcluidoComAvisos
	default:
		r.Estado = Concluido
	}
	return r
}

// rodarSeguro transforma um panic num módulo em erro, para um bug nele não
// derrubar a interface inteira.
func rodarSeguro(s *sistema.Sistema, m *Modulo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("erro interno: %v", r)
		}
	}()
	return m.Executar(s)
}
