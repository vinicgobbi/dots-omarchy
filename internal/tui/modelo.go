// Package tui é a interface interativa do setup: escolha do usuário-alvo e do
// tema, checklist de módulos com dependências, revisão, execução com progresso
// e saída ao vivo, e o resumo final.
package tui

import (
	"context"
	"fmt"
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
	Notas       []string // efeitos do --pular, mostrados na revisão
	FalhaRapida bool     // --fail-fast
	Trabalho    string   // diretório temporário da execução
	// TemaAtivo é a URL do tema escolhido; com TemaEscolhido (veio de --tema),
	// não pergunta.
	TemaAtivo     string
	TemaEscolhido bool
}

// Resultado é o que aconteceu na sessão, para o main.
type Resultado struct {
	Iniciado bool
	Abortado bool
	Modulos  []modulos.Resultado
	Duracao  time.Duration
}

type tela int

const (
	telaUsuario tela = iota
	telaOutroUsuario
	telaTema
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
	loteMsg         []linha
	moduloInicioMsg struct{ n int }
	moduloFimMsg    struct {
		n int
		r modulos.Resultado
	}
	fimMsg       struct{}
	confirmarMsg struct {
		c        ui.Confirmacao
		resposta chan bool
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

	// Tema
	cursorTema int

	// Módulos
	sel    []bool
	cursor int
	nota   string
	notas  []string // dependências puxadas do --modulos e efeitos do --pular

	// Execução
	fila         []*modulos.Modulo
	resultados   []modulos.Resultado
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
	notas := append([]string(nil), cfg.Notas...)
	for _, p := range modulos.Resolver(cfg.Modulos, sel) {
		notas = append(notas, fmt.Sprintf("'%s' depende de '%s': selecionado automaticamente.",
			cfg.Modulos[p.Por].Titulo, cfg.Modulos[p.Dep].Titulo))
	}

	m := modelo{
		cfg: cfg, ponte: b, largura: 80, altura: 24,
		entrada: entrada, spinner: sp, sel: sel, notas: notas,
		voce: sistema.DetectarUsuario(),
	}
	m.contas = sistema.ContasHumanas()
	for i, c := range m.contas {
		if c.Nome == m.voce {
			m.cursorConta = i
		}
	}
	if cfg.Dados != nil {
		m.cursorTema = max(0, len(cfg.Dados.Themes)-1)
		for i, url := range cfg.Dados.Themes {
			if url == cfg.TemaAtivo {
				m.cursorTema = i
			}
		}
	}

	if cfg.Conta != nil {
		m.conta, m.contaDaFlag = *cfg.Conta, true
		m.tela = m.telaAposUsuario()
	}
	return m
}

func (m modelo) marcado(id string) bool {
	for i, mod := range m.cfg.Modulos {
		if mod.ID == id {
			return m.sel[i]
		}
	}
	return false
}

// perguntaTema diz se a tela de tema se aplica: há temas, o módulo está marcado
// e a escolha não veio de --tema.
func (m modelo) perguntaTema() bool {
	return !m.cfg.TemaEscolhido && m.cfg.Dados != nil && len(m.cfg.Dados.Themes) > 0 && m.marcado("temas_omarchy")
}

func (m modelo) telaAposUsuario() tela {
	if m.perguntaTema() {
		return telaTema
	}
	return m.telaAposTema()
}

func (m modelo) telaAposTema() tela {
	if m.cfg.PularMenu {
		return telaRevisao
	}
	return telaModulos
}

func (m modelo) Init() tea.Cmd { return m.spinner.Tick }

func (m modelo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Descontando a margem aplicada em View.
		m.largura, m.altura = max(20, msg.Width-margemLateral), max(10, msg.Height-margemTopo)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case loteMsg:
		for _, l := range msg {
			m.adicionar(l)
		}
		return m, nil

	case moduloInicioMsg:
		m.atual, m.inicioModulo = msg.n, time.Now()
		m.resultados[msg.n].Estado = modulos.Rodando
		m.adicionar(linha{linhaPasso, fmt.Sprintf("[%d/%d] %s", msg.n+1, len(m.fila), m.fila[msg.n].Titulo)})
		return m, nil

	case moduloFimMsg:
		m.resultados[msg.n] = msg.r
		if msg.r.Estado == modulos.Falhou && !m.abortado {
			m.adicionar(linha{linhaErro, "Falhou: " + msg.r.Erro.Error()})
		}
		return m, nil

	case fimMsg:
		m.terminou, m.duracao = true, time.Since(m.inicio)
		for i, r := range m.resultados {
			if r.Estado == modulos.Pendente || r.Estado == modulos.Rodando {
				m.resultados[i].Estado = modulos.Cancelado
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
	case telaTema:
		return m.teclaTema(k)
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
	m.pedirAbortar = false
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

// teclaTema: a última opção da lista é não instalar temas extras.
func (m modelo) teclaTema(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	temas := m.cfg.Dados.Themes
	switch k.String() {
	case "up", "k":
		m.cursorTema = max(0, m.cursorTema-1)
	case "down", "j":
		m.cursorTema = min(len(temas), m.cursorTema+1)
	case "enter", " ":
		m.sel = append([]bool(nil), m.sel...)
		if m.cursorTema < len(temas) {
			m.cfg.TemaAtivo = temas[m.cursorTema]
			m.tela = m.telaAposTema()
			break
		}
		// Sem tema, o módulo não teria o que fazer: desmarca.
		m.cfg.TemaAtivo = ""
		for i, mod := range m.cfg.Modulos {
			if mod.ID == "temas_omarchy" {
				modulos.Desmarcar(m.cfg.Modulos, m.sel, i)
			}
		}
		m.tela = m.telaAposTema()
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
		switch {
		case m.perguntaTema():
			m.tela = telaTema
		case !m.contaDaFlag:
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
	var nomes []string
	if m.sel[i] {
		for _, j := range modulos.Desmarcar(lista, m.sel, i) {
			nomes = append(nomes, lista[j].Titulo)
		}
		if len(nomes) == 0 {
			return ""
		}
		return "Também desmarcado (depende deste): " + strings.Join(nomes, ", ") + "."
	}
	for _, p := range modulos.Marcar(lista, m.sel, i) {
		nomes = append(nomes, lista[p.Dep].Titulo)
	}
	if len(nomes) == 0 {
		return ""
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
	case "enter", "s", "S", "y", "Y":
		return m.iniciar()
	case "esc", "n", "N", "backspace":
		m.tela = telaModulos
	case "q":
		m.abortado = true
		return m, tea.Quit
	}
	return m, nil
}

func (m modelo) teclaExecucao(k tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "n":
		m.pedirAbortar = false
	case "esc":
		m.pedirAbortar = false
		if m.terminou {
			m.tela = telaResumo
		}
	case "enter", "r":
		if m.terminou {
			m.tela = telaResumo
		}
	case "q":
		if m.terminou {
			return m, tea.Quit
		}
		return m.ctrlC()
	}
	return m, nil
}

func (m modelo) iniciar() (tea.Model, tea.Cmd) {
	for i, mod := range m.cfg.Modulos {
		if m.sel[i] {
			m.fila = append(m.fila, mod)
		}
	}
	m.resultados = make([]modulos.Resultado, len(m.fila))
	for i, mod := range m.fila {
		m.resultados[i].Modulo = mod
	}
	m.tela, m.inicio = telaExecucao, time.Now()

	ctx, cancelar := context.WithCancel(context.Background())
	m.cancelar = cancelar

	log := m.cfg.Log
	if err := log.IniciarBruto(); err != nil {
		m.adicionar(linha{linhaAviso, "Não foi possível abrir o log bruto: " + err.Error()})
	}
	if !m.cfg.DryRun {
		if err := log.Chown(m.conta.UID, m.conta.GID); err != nil {
			m.adicionar(linha{linhaAviso, fmt.Sprintf("Não foi possível passar os logs para %s: %v", m.conta.Nome, err)})
		}
	}
	log.Linha("INFO", "Configuração será aplicada para: "+m.conta.Nome)
	for _, mod := range m.fila {
		log.Linha("INFO", "Módulo selecionado: "+mod.Titulo)
	}
	if m.cfg.DryRun {
		m.adicionar(linha{linhaAviso, "Modo --dry-run: nada será alterado no sistema."})
	}

	s := sistema.Novo(ctx, m.cfg.Raiz, m.conta, m.cfg.Dados, m.cfg.DryRun, m.ponte, log)
	s.Trabalho, s.TemaAtivo = m.cfg.Trabalho, m.cfg.TemaAtivo
	go executar(ctx, s, m.fila, m.cfg.FalhaRapida, m.ponte)
	return m, nil
}

// executar roda os módulos numa goroutine, mandando o andamento para a tela.
func executar(ctx context.Context, s *sistema.Sistema, fila []*modulos.Modulo, falhaRapida bool, b *ponte) {
	defer b.enviar(fimMsg{})
	restaurar, err := s.LiberarSudo()
	if err != nil {
		b.enviar(moduloInicioMsg{0})
		b.enviar(moduloFimMsg{0, modulos.Resultado{Modulo: fila[0], Estado: modulos.Falhou, Erro: err}})
		return
	}
	defer restaurar()

	modulos.Rodar(ctx, s, fila, falhaRapida, modulos.Observador{
		Inicio: func(n int) {
			b.enviar(moduloInicioMsg{n})
			passo := fmt.Sprintf("[%d/%d] %s", n+1, len(fila), fila[n].Titulo)
			b.log.Linha("PASSO", passo)
			fmt.Fprintln(b.log, passo)
		},
		Fim: func(n int, r modulos.Resultado) {
			if r.Estado == modulos.Falhou {
				b.log.Linha("ERRO", r.Erro.Error())
				fmt.Fprintf(b.log, "[-] %s\n", r.Erro)
			}
			b.enviar(moduloFimMsg{n, r})
		},
	})
}

func (m modelo) resultado() Resultado {
	return Resultado{Iniciado: m.fila != nil, Abortado: m.abortado,
		Modulos: m.resultados, Duracao: m.duracao}
}

// Rodar abre a TUI e devolve o que aconteceu.
func Rodar(cfg Config) (Resultado, error) {
	b := novaPonte(cfg.Log)
	p := tea.NewProgram(novoModelo(cfg, b), tea.WithAltScreen())
	b.despachar = p.Send
	final, err := p.Run()
	if err != nil {
		return Resultado{}, err
	}
	return final.(modelo).resultado(), nil
}
