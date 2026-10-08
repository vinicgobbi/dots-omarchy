// Package tui é a interface interativa do setup: escolha do usuário-alvo,
// checklist de módulos com dependências, revisão, execução com progresso e
// saída ao vivo, e o resumo final.
package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

type Config struct {
	Versao      string
	Raiz        string
	Dados       *dados.Dados
	DryRun      bool
	Log         *logs.Log
	NomeSistema string
	Modulos     []*modulos.Modulo
	Selecao     []bool // pré-seleção; nil marca tudo
	PularMenu   bool   // a seleção veio de --modulos
	Conta       *sistema.Conta
}

type Estado int

const (
	Pendente Estado = iota
	Rodando
	Concluido
	Falhou
	NaoExecutado
)

type ResumoModulo struct {
	Titulo  string
	Estado  Estado
	Duracao time.Duration
	Erro    error
}

type Resultado struct {
	Iniciado bool
	Abortado bool
	Falha    error
	Modulos  []ResumoModulo
	Avisos   []string
	Duracao  time.Duration
}

type tela int

const (
	telaUsuario tela = iota
	telaOutroUsuario
	telaModulos
	telaRevisao
	telaExecucao
	telaResumo
)

type tipoLinha int

const (
	linhaSaida tipoLinha = iota
	linhaInfo
	linhaSucesso
	linhaAviso
	linhaErro
	linhaDryRun
	linhaPasso
)

type linha struct {
	tipo  tipoLinha
	texto string
}

// Mensagens da goroutine que roda os módulos.
type (
	linhaMsg        linha
	moduloInicioMsg struct{ n int }
	moduloFimMsg    struct {
		n       int
		err     error
		duracao time.Duration
	}
	fimMsg       struct{}
	confirmarMsg struct {
		c        ui.Confirmacao
		resposta chan bool
	}
	execMsg struct {
		cmd *exec.Cmd
		fim chan error
	}
	execFimMsg struct {
		err error
		fim chan error
	}
)

const maxLinhas = 3000

type modelo struct {
	cfg     Config
	ponte   *ponte
	largura int
	altura  int
	tela    tela

	// Usuário
	contas      []sistema.Conta
	voce        string
	cursorConta int
	entrada     textinput.Model
	erroEntrada string
	conta       sistema.Conta
	contaDaFlag bool

	// Módulos
	sel    []bool
	cursor int
	nota   string
	notas  []string // dependências puxadas do --modulos

	// Execução
	fila         []*modulos.Modulo
	estados      []Estado
	duracoes     []time.Duration
	erros        []error
	atual        int
	inicioModulo time.Time
	linhas       []linha
	rolagem      int // linhas acima do fim; 0 = seguindo a saída
	spinner      spinner.Model
	inicio       time.Time
	duracao      time.Duration
	modal        *confirmarMsg
	modalSim     bool
	modalTopo    int
	pedirAbortar bool
	abortado     bool
	terminou     bool
	avisos       []string
	cancelar     context.CancelFunc
}

func novoModelo(cfg Config, b *ponte) modelo {
	entrada := textinput.New()
	entrada.Placeholder = "nome do usuário"
	entrada.CharLimit = 64
	entrada.Prompt = "› "

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = estiloDestaque

	sel := cfg.Selecao
	if sel == nil {
		sel = make([]bool, len(cfg.Modulos))
		for i := range sel {
			sel[i] = true
		}
	}
	var notas []string
	for _, p := range modulos.Resolver(cfg.Modulos, sel) {
		notas = append(notas, fmt.Sprintf("'%s' depende de '%s': selecionado automaticamente.",
			cfg.Modulos[p.Por].Titulo, cfg.Modulos[p.Dep].Titulo))
	}

	m := modelo{
		cfg: cfg, ponte: b, largura: 80, altura: 24,
		entrada: entrada, spinner: sp, sel: sel, notas: notas,
		voce: os.Getenv("SUDO_USER"),
	}
	if m.voce == "" {
		m.voce = os.Getenv("USER")
	}
	m.contas = sistema.ContasHumanas()
	for i, c := range m.contas {
		if c.Nome == m.voce {
			m.cursorConta = i
		}
	}

	if cfg.Conta != nil {
		m.conta, m.contaDaFlag = *cfg.Conta, true
		m.tela = m.telaAposUsuario()
	}
	return m
}

func (m modelo) telaAposUsuario() tela {
	if m.cfg.PularMenu {
		return telaRevisao
	}
	return telaModulos
}

func (m modelo) Init() tea.Cmd { return m.spinner.Tick }

func (m modelo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.largura, m.altura = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case linhaMsg:
		m.adicionar(linha(msg))
		if msg.tipo == linhaAviso && m.atual < len(m.fila) {
			m.avisos = append(m.avisos, m.fila[m.atual].Titulo+": "+msg.texto)
		}
		return m, nil

	case moduloInicioMsg:
		m.atual, m.inicioModulo = msg.n, time.Now()
		m.estados[msg.n] = Rodando
		m.adicionar(linha{linhaPasso, fmt.Sprintf("[%d/%d] %s", msg.n+1, len(m.fila), m.fila[msg.n].Titulo)})
		return m, nil

	case moduloFimMsg:
		m.duracoes[msg.n] = msg.duracao
		if msg.err != nil {
			m.estados[msg.n], m.erros[msg.n] = Falhou, msg.err
			if !m.abortado {
				m.adicionar(linha{linhaErro, "Falhou: " + msg.err.Error()})
			}
		} else {
			m.estados[msg.n] = Concluido
		}
		return m, nil

	case fimMsg:
		m.terminou, m.duracao = true, time.Since(m.inicio)
		for i, e := range m.estados {
			if e == Pendente {
				m.estados[i] = NaoExecutado
			}
		}
		if m.abortado {
			return m, tea.Quit
		}
		m.tela = telaResumo
		return m, nil

	case confirmarMsg:
		if m.abortado {
			msg.resposta <- false
			return m, nil
		}
		m.modal, m.modalSim, m.modalTopo = &msg, false, 0
		return m, nil

	case execMsg:
		titulo := m.fila[m.atual].Titulo
		return m, tea.ExecProcess(comAviso(msg.cmd, titulo), func(err error) tea.Msg {
			return execFimMsg{err, msg.fim}
		})

	case execFimMsg:
		msg.fim <- msg.err
		return m, nil

	case tea.KeyMsg:
		return m.tecla(msg)
	}

	if m.tela == telaOutroUsuario {
		var cmd tea.Cmd
		m.entrada, cmd = m.entrada.Update(msg)
		return m, cmd
	}
	return m, nil
}

// comAviso mostra, antes de um comando interativo, o que está acontecendo e
// por que a TUI sumiu por um momento.
func comAviso(cmd *exec.Cmd, titulo string) *exec.Cmd {
	script := `printf '\n\033[1;33m[*] %s\033[0m\n\033[1;30m%s\033[0m\n\n' "$1" "$2"; shift 2; exec "$@"`
	args := append([]string{"-c", script, "bash", titulo,
		"Este passo usa o terminal diretamente (pode pedir sua senha). A tela volta sozinha ao terminar."},
		cmd.Args...)
	novo := exec.Command("bash", args...)
	novo.Env = cmd.Env
	return novo
}

func (m *modelo) adicionar(l linha) {
	m.linhas = append(m.linhas, l)
	if len(m.linhas) > maxLinhas {
		m.linhas = m.linhas[len(m.linhas)-maxLinhas:]
	}
	if m.rolagem > 0 {
		m.rolagem++ // mantém o trecho que o usuário está lendo parado
	}
}

func (m modelo) tecla(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k.String() == "ctrl+c" {
		return m.ctrlC()
	}
	if m.modal != nil {
		return m.teclaModal(k), nil
	}
	switch m.tela {
	case telaUsuario:
		return m.teclaUsuario(k)
	case telaOutroUsuario:
		return m.teclaOutroUsuario(k)
	case telaModulos:
		return m.teclaModulos(k)
	case telaRevisao:
		return m.teclaRevisao(k)
	case telaExecucao:
		return m.teclaExecucao(k)
	case telaResumo:
		switch k.String() {
		case "q", "esc", "enter":
			return m, tea.Quit
		case "l", "s":
			m.tela = telaExecucao
		}
	}
	return m, nil
}

// ctrlC: antes de começar, sai na hora; durante a execução, pede uma segunda
// vez e então interrompe o comando atual e para.
func (m modelo) ctrlC() (tea.Model, tea.Cmd) {
	if m.tela != telaExecucao || m.terminou {
		if m.tela < telaExecucao {
			m.abortado = true
		}
		return m, tea.Quit
	}
	if m.abortado {
		// Terceira vez: algo não respondeu ao cancelamento.
		return m, tea.Quit
	}
	if !m.pedirAbortar {
		m.pedirAbortar = true
		return m, nil
	}
	m.abortado = true
	m.cancelar()
	if m.modal != nil {
		m.modal.resposta <- false
		m.modal = nil
	}
	m.adicionar(linha{linhaErro, "Abortando a pedido do usuário..."})
	return m, nil
}

func (m modelo) teclaModal(k tea.KeyMsg) modelo {
	responder := func(sim bool) modelo {
		m.modal.resposta <- sim
		m.modal = nil
		return m
	}
	corpo, cabe := m.corpoModal()
	switch k.String() {
	case "up", "k":
		m.modalTopo = max(0, m.modalTopo-1)
	case "down", "j":
		m.modalTopo = max(0, min(len(corpo)-cabe, m.modalTopo+1))
	case "left", "right", "tab", "shift+tab", "h", "l":
		m.modalSim = !m.modalSim
	case "s", "S", "y", "Y":
		return responder(true)
	case "n", "N", "esc":
		return responder(false)
	case "enter":
		return responder(m.modalSim)
	}
	return m
}

func (m modelo) totalContas() int { return len(m.contas) + 1 } // + "Outro usuário…"

func (m modelo) teclaUsuario(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "up", "k":
		m.cursorConta = max(0, m.cursorConta-1)
	case "down", "j":
		m.cursorConta = min(m.totalContas()-1, m.cursorConta+1)
	case "enter":
		if m.cursorConta == len(m.contas) {
			m.tela, m.erroEntrada = telaOutroUsuario, ""
			m.entrada.SetValue("")
			return m, m.entrada.Focus()
		}
		m.conta = m.contas[m.cursorConta]
		m.tela = m.telaAposUsuario()
	case "q", "esc":
		m.abortado = true
		return m, tea.Quit
	}
	return m, nil
}

func (m modelo) teclaOutroUsuario(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.entrada.Blur()
		m.tela = telaUsuario
		return m, nil
	case "enter":
		nome := strings.TrimSpace(m.entrada.Value())
		if nome == "" {
			m.erroEntrada = "Digite um nome de usuário."
			return m, nil
		}
		conta, err := sistema.BuscarConta(nome)
		if err != nil {
			m.erroEntrada = err.Error() + ". Tente novamente."
			return m, nil
		}
		m.entrada.Blur()
		m.conta = conta
		m.tela = m.telaAposUsuario()
		return m, nil
	}
	var cmd tea.Cmd
	m.entrada, cmd = m.entrada.Update(k)
	m.erroEntrada = ""
	return m, cmd
}

func (m modelo) teclaModulos(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	lista := m.cfg.Modulos
	switch k.String() {
	case "up", "k":
		m.cursor = max(0, m.cursor-1)
	case "down", "j":
		m.cursor = min(len(lista)-1, m.cursor+1)
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(lista) - 1
	case "pgup":
		m.cursor = max(0, m.cursor-10)
	case "pgdown":
		m.cursor = min(len(lista)-1, m.cursor+10)
	case " ", "x":
		m.sel = append([]bool(nil), m.sel...)
		m.nota = m.alternar(m.cursor)
	case "a":
		todos := true
		for _, s := range m.sel {
			todos = todos && s
		}
		m.sel = make([]bool, len(lista))
		for i := range m.sel {
			m.sel[i] = !todos
		}
		m.nota = ""
	case "enter":
		if m.selecionados() == 0 {
			m.nota = "Selecione pelo menos um módulo."
			return m, nil
		}
		m.nota, m.tela = "", telaRevisao
	case "esc":
		if !m.contaDaFlag {
			m.tela = telaUsuario
		}
	case "q":
		m.abortado = true
		return m, tea.Quit
	}
	return m, nil
}

// alternar marca/desmarca o módulo i mantendo as dependências consistentes e
// devolve a explicação do que mudou além dele.
func (m modelo) alternar(i int) string {
	lista := m.cfg.Modulos
	if m.sel[i] {
		if deps := modulos.Dependentes(lista, m.sel, i); len(deps) > 0 {
			var nomes []string
			for _, j := range deps {
				nomes = append(nomes, lista[j].Titulo)
			}
			return fmt.Sprintf("'%s' é necessário para: %s. Desmarque esses antes.",
				lista[i].Titulo, strings.Join(nomes, ", "))
		}
		m.sel[i] = false
		return ""
	}
	puxados := modulos.Marcar(lista, m.sel, i)
	if len(puxados) == 0 {
		return ""
	}
	var nomes []string
	for _, p := range puxados {
		nomes = append(nomes, lista[p.Dep].Titulo)
	}
	return "Dependência marcada junto: " + strings.Join(nomes, ", ") + "."
}

func (m modelo) selecionados() int {
	n := 0
	for _, s := range m.sel {
		if s {
			n++
		}
	}
	return n
}

func (m modelo) teclaRevisao(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "enter":
		return m.iniciar()
	case "esc":
		m.tela = telaModulos
	case "q":
		m.abortado = true
		return m, tea.Quit
	}
	return m, nil
}

func (m modelo) teclaExecucao(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.pedirAbortar = false
	maxRolagem := max(0, len(m.linhas)-1)
	switch k.String() {
	case "up", "k":
		m.rolagem = min(maxRolagem, m.rolagem+1)
	case "down", "j":
		m.rolagem = max(0, m.rolagem-1)
	case "pgup":
		m.rolagem = min(maxRolagem, m.rolagem+m.alturaSaida())
	case "pgdown":
		m.rolagem = max(0, m.rolagem-m.alturaSaida())
	case "home", "g":
		m.rolagem = maxRolagem
	case "end", "G", "f":
		m.rolagem = 0
	case "enter", "esc", "r":
		if m.terminou {
			m.tela = telaResumo
		}
	case "q":
		if m.terminou {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m modelo) iniciar() (tea.Model, tea.Cmd) {
	for i, mod := range m.cfg.Modulos {
		if m.sel[i] {
			m.fila = append(m.fila, mod)
		}
	}
	n := len(m.fila)
	m.estados = make([]Estado, n)
	m.duracoes = make([]time.Duration, n)
	m.erros = make([]error, n)
	m.tela, m.inicio = telaExecucao, time.Now()

	ctx, cancelar := context.WithCancel(context.Background())
	m.cancelar = cancelar

	log := m.cfg.Log
	if err := log.IniciarBruto(); err != nil {
		m.adicionar(linha{linhaAviso, "Não foi possível abrir o log bruto: " + err.Error()})
	}
	log.Linha("INFO", "Configuração será aplicada para: "+m.conta.Nome)
	for _, mod := range m.fila {
		log.Linha("INFO", "Módulo selecionado: "+mod.Titulo)
	}
	if m.cfg.DryRun {
		m.adicionar(linha{linhaAviso, "Modo --dry-run: nada será alterado no sistema."})
	}

	s := sistema.Novo(ctx, m.cfg.Raiz, m.conta, m.cfg.Dados, m.cfg.DryRun, m.ponte, log)
	go executar(s, m.fila, m.ponte)
	return m, nil
}

// executar roda os módulos em ordem; o primeiro que falhar interrompe o resto
// (como o set -e da versão em bash).
func executar(s *sistema.Sistema, fila []*modulos.Modulo, b *ponte) {
	for n, mod := range fila {
		b.enviar(moduloInicioMsg{n})
		passo := fmt.Sprintf("[%d/%d] %s", n+1, len(fila), mod.Titulo)
		b.log.Linha("PASSO", passo)
		fmt.Fprintln(b.log, passo)

		inicio := time.Now()
		err := rodarModulo(s, mod)
		if err != nil {
			b.log.Linha("ERRO", err.Error())
			fmt.Fprintf(b.log, "[-] %s\n", err)
		}
		b.enviar(moduloFimMsg{n, err, time.Since(inicio)})
		if err != nil {
			break
		}
	}
	b.enviar(fimMsg{})
}

func rodarModulo(s *sistema.Sistema, mod *modulos.Modulo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("erro interno: %v", r)
		}
	}()
	return mod.Executar(s)
}

func (m modelo) resultado() Resultado {
	r := Resultado{Iniciado: m.fila != nil, Abortado: m.abortado, Avisos: m.avisos, Duracao: m.duracao}
	for i, mod := range m.fila {
		r.Modulos = append(r.Modulos, ResumoModulo{mod.Titulo, m.estados[i], m.duracoes[i], m.erros[i]})
		if m.erros[i] != nil && r.Falha == nil {
			r.Falha = m.erros[i]
		}
	}
	return r
}

// Rodar abre a TUI e devolve o que aconteceu.
func Rodar(cfg Config) (Resultado, error) {
	b := novaPonte(cfg.Log)
	p := tea.NewProgram(novoModelo(cfg, b), tea.WithAltScreen())
	b.p = p
	final, err := p.Run()
	if err != nil {
		return Resultado{}, err
	}
	return final.(modelo).resultado(), nil
}
