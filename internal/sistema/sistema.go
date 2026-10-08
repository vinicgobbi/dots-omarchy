// Package sistema executa comandos e escreve arquivos, como root ou como o
// usuário-alvo. Tudo que altera o sistema passa por aqui, para o --dry-run
// conseguir só mostrar o que seria feito.
package sistema

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// FalhaComando é um comando que terminou com código diferente de zero
// (equivale ao set -e).
type FalhaComando struct {
	Comando string
	Codigo  int
}

func (f *FalhaComando) Error() string {
	return fmt.Sprintf("%s (código %d)", f.Comando, f.Codigo)
}

// Opts ajusta Rodar.
type Opts struct {
	SemCheck bool // não falha com código diferente de zero
	Quieto   bool // descarta stdout/stderr (&>/dev/null)
	Capturar bool // devolve o stdout em Resultado.Saida
	Leitura  bool // comando só de consulta: roda mesmo no --dry-run
	Espelhar bool // mostra a saída e também devolve stdout+stderr em Resultado.Saida
}

type Resultado struct {
	Codigo int
	Saida  string
}

type Sistema struct {
	Raiz   string
	Conta  Conta
	Dados  *dados.Dados
	DryRun bool
	UI     ui.UI
	Log    *logs.Log
	ctx    context.Context
}

func Novo(ctx context.Context, raiz string, conta Conta, d *dados.Dados, dryRun bool, u ui.UI, log *logs.Log) *Sistema {
	return &Sistema{Raiz: raiz, Conta: conta, Dados: d, DryRun: dryRun, UI: u, Log: log, ctx: ctx}
}

func (s *Sistema) Usuario() string { return s.Conta.Nome }
func (s *Sistema) Home() string    { return s.Conta.Home }

// ------------------------------------------------------------------
// Comandos
// ------------------------------------------------------------------

// Rodar roda um comando como root.
func (s *Sistema) Rodar(args []string, o Opts) (Resultado, error) {
	texto := Juntar(args...)
	if s.DryRun && !o.Leitura {
		s.UI.DryRun(texto)
		return Resultado{}, nil
	}
	if err := s.ctx.Err(); err != nil {
		return Resultado{Codigo: -1}, err
	}

	res := s.rodarCapturado(args, o)
	if err := s.ctx.Err(); err != nil {
		return res, err
	}
	if !o.SemCheck && res.Codigo != 0 {
		return res, &FalhaComando{texto, res.Codigo}
	}
	return res, nil
}

func (s *Sistema) rodarCapturado(args []string, o Opts) Resultado {
	cmd := exec.CommandContext(s.ctx, args[0], args[1:]...)
	var captura bytes.Buffer
	saida := s.UI.Saida()
	switch {
	case o.Capturar:
		cmd.Stdout = &captura
		if !o.Quieto {
			cmd.Stderr = saida
		}
	case o.Espelhar:
		espelho := io.MultiWriter(saida, &captura)
		cmd.Stdout, cmd.Stderr = espelho, espelho
	case !o.Quieto:
		cmd.Stdout, cmd.Stderr = saida, saida
	}
	if s.UI.Isolado() {
		// Sessão própria: sem terminal de controle, um prompt inesperado
		// falha em vez de disputar a tela com a TUI. O cancelamento
		// mata o grupo todo (su, bash e o que mais tiver sido aberto).
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM) }
	}

	err := cmd.Run()
	if f, ok := saida.(interface{ Flush() }); ok && !o.Quieto {
		f.Flush()
	}
	return Resultado{Codigo: codigoSaida(err), Saida: captura.String()}
}

func codigoSaida(err error) int {
	var saida *exec.ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &saida):
		if saida.ExitCode() < 0 {
			return 128 + int(syscall.SIGTERM)
		}
		return saida.ExitCode()
	default:
		// Comando inexistente ou que não pôde ser iniciado, como no shell.
		return 127
	}
}

// Run roda um comando como root e falha se ele falhar.
func (s *Sistema) Run(args ...string) error {
	_, err := s.Rodar(args, Opts{})
	return err
}

// ComoUsuario roda como o usuário-alvo em login shell (su -), com $HOME
// correto. Se já somos esse usuário (--dry-run sem sudo), usa um login shell
// direto.
func (s *Sistema) ComoUsuario(script string, o Opts) (Resultado, error) {
	if os.Geteuid() == s.Conta.UID {
		return s.Rodar([]string{"bash", "-lc", script}, o)
	}
	return s.Rodar([]string{"su", "-", s.Conta.Nome, "-c", script}, o)
}

// Usr roda como o usuário-alvo e falha se o comando falhar.
func (s *Sistema) Usr(script string) error {
	_, err := s.ComoUsuario(script, Opts{})
	return err
}

// Ok é uma consulta silenciosa: true se o comando terminar com sucesso.
func (s *Sistema) Ok(args ...string) bool {
	res, err := s.Rodar(args, Opts{SemCheck: true, Quieto: true, Leitura: true})
	return err == nil && res.Codigo == 0
}

// OkUsuario é o Ok como o usuário-alvo.
func (s *Sistema) OkUsuario(script string) bool {
	res, err := s.ComoUsuario(script, Opts{SemCheck: true, Quieto: true, Leitura: true})
	return err == nil && res.Codigo == 0
}

// Tentar roda como o usuário-alvo sem falhar o módulo; devolve false (e o
// erro de cancelamento, se houver) quando não deu certo.
func (s *Sistema) Tentar(script string, o Opts) (bool, error) {
	o.SemCheck = true
	res, err := s.ComoUsuario(script, o)
	return err == nil && res.Codigo == 0, err
}

func (s *Sistema) Pacman(pacotes ...string) error {
	if len(pacotes) == 0 {
		return nil
	}
	return s.Run(append([]string{"pacman", "-S", "--needed", "--noconfirm"}, pacotes...)...)
}

// Aur instala via yay como o usuário-alvo (o makepkg recusa rodar como root).
// Depende do módulo "repositorios" ter liberado sudo sem senha para o pacman
// nesse usuário antes.
func (s *Sistema) Aur(pacotes ...string) error {
	if len(pacotes) == 0 {
		return nil
	}
	return s.Usr(Juntar(append([]string{"yay", "-S", "--needed", "--noconfirm"}, pacotes...)...))
}

func (s *Sistema) Flatpak(apps ...string) error {
	if err := s.Run("flatpak", "remote-add", "--if-not-exists", "flathub",
		"https://flathub.org/repo/flathub.flatpakrepo"); err != nil {
		return err
	}
	if len(apps) == 0 {
		return nil
	}
	return s.Run(append([]string{"flatpak", "install", "-y", "flathub"}, apps...)...)
}

func ComandoExiste(nome string) bool {
	_, err := exec.LookPath(nome)
	return err == nil
}

func GrupoExiste(nome string) bool {
	return exec.Command("getent", "group", nome).Run() == nil
}

// ------------------------------------------------------------------
// Contas
// ------------------------------------------------------------------

type Conta struct {
	Nome  string
	UID   int
	GID   int
	Home  string
	Shell string
	Gecos string
}

func parseConta(linha string) (Conta, bool) {
	campos := strings.Split(strings.TrimSpace(linha), ":")
	if len(campos) < 7 {
		return Conta{}, false
	}
	uid, err1 := strconv.Atoi(campos[2])
	gid, err2 := strconv.Atoi(campos[3])
	if err1 != nil || err2 != nil {
		return Conta{}, false
	}
	return Conta{Nome: campos[0], UID: uid, GID: gid, Gecos: strings.Split(campos[4], ",")[0],
		Home: campos[5], Shell: campos[6]}, true
}

// BuscarConta resolve via NSS (getent), o que inclui contas de domínio
// (AD/LDAP via sssd/winbind).
func BuscarConta(nome string) (Conta, error) {
	saida, err := exec.Command("getent", "passwd", nome).Output()
	if err != nil {
		return Conta{}, fmt.Errorf("usuário '%s' não existe neste sistema", nome)
	}
	conta, ok := parseConta(string(saida))
	if !ok {
		return Conta{}, fmt.Errorf("não foi possível ler a conta '%s'", nome)
	}
	return conta, nil
}

// ContasHumanas lista as contas de gente (UID >= 1000, com shell de login).
func ContasHumanas() []Conta {
	saida, _ := exec.Command("getent", "passwd").Output()
	var contas []Conta
	for _, linha := range strings.Split(string(saida), "\n") {
		conta, ok := parseConta(linha)
		if !ok || conta.UID < 1000 || conta.UID >= 60000 ||
			strings.HasSuffix(conta.Shell, "nologin") || strings.HasSuffix(conta.Shell, "false") {
			continue
		}
		contas = append(contas, conta)
	}
	return contas
}

// ------------------------------------------------------------------
// Utilitários
// ------------------------------------------------------------------

var seguro = regexp.MustCompile(`^[A-Za-z0-9@%+=:,./_-]+$`)

// Juntar monta uma linha de shell com cada argumento citado quando preciso
// (equivale ao shlex.join).
func Juntar(args ...string) string {
	citados := make([]string, len(args))
	for i, a := range args {
		if seguro.MatchString(a) {
			citados[i] = a
		} else {
			citados[i] = "'" + strings.ReplaceAll(a, "'", `'"'"'`) + "'"
		}
	}
	return strings.Join(citados, " ")
}
