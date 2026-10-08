package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
)

// Só cores ANSI básicas (0–15): seguem a paleta do tema do Omarchy aplicada
// no terminal.
var (
	corVerde    = lipgloss.Color("2")
	corAmarelo  = lipgloss.Color("3")
	corAzul     = lipgloss.Color("4")
	corVermelho = lipgloss.Color("1")
	corMagenta  = lipgloss.Color("5")
	corCiano    = lipgloss.Color("6")
	corCinza    = lipgloss.Color("8")

	estiloMarca    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(corVerde).Padding(0, 1)
	estiloNegrito  = lipgloss.NewStyle().Bold(true)
	estiloSutil    = lipgloss.NewStyle().Foreground(corCinza)
	estiloDestaque = lipgloss.NewStyle().Foreground(corVerde).Bold(true)
	estiloAviso    = lipgloss.NewStyle().Foreground(corAmarelo)
	estiloErro     = lipgloss.NewStyle().Foreground(corVermelho)
	estiloInfo     = lipgloss.NewStyle().Foreground(corAzul)
	estiloDryRun   = lipgloss.NewStyle().Foreground(corMagenta)
	estiloPasso    = lipgloss.NewStyle().Foreground(corCiano).Bold(true)
	estiloSelo     = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(corMagenta).Bold(true).Padding(0, 1)
	estiloCaixa    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(corCinza).Padding(0, 1)
)

func (m modelo) View() string {
	if m.modal != nil {
		return m.viewModal()
	}
	var corpo, rodape string
	switch m.tela {
	case telaUsuario:
		corpo, rodape = m.viewUsuario(), ajuda("↑/↓", "mover", "enter", "escolher", "q", "sair")
	case telaOutroUsuario:
		corpo, rodape = m.viewOutroUsuario(), ajuda("enter", "confirmar", "esc", "voltar")
	case telaModulos:
		corpo = m.viewModulos()
		rodape = ajuda("↑/↓", "mover", "espaço", "marcar", "a", "todos", "enter", "revisar", "esc", "voltar", "q", "sair")
	case telaRevisao:
		corpo, rodape = m.viewRevisao(), ajuda("enter", "começar", "esc", "voltar", "q", "sair")
	case telaExecucao:
		corpo = m.viewExecucao()
		switch {
		case m.terminou:
			rodape = ajuda("↑/↓ pgup/pgdn", "rolar", "enter", "resumo", "q", "sair")
		case m.abortado:
			rodape = estiloErro.Render("Abortando... (ctrl+c de novo força a saída)")
		case m.pedirAbortar:
			rodape = estiloErro.Bold(true).Render("Pressione ctrl+c de novo para abortar a instalação.")
		default:
			rodape = ajuda("↑/↓ pgup/pgdn", "rolar a saída", "end", "seguir", "ctrl+c", "abortar")
		}
	case telaResumo:
		corpo, rodape = m.viewResumo(), ajuda("l", "ver a saída", "q", "sair")
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.cabecalho(), "", corpo) +
		"\n" + strings.Repeat("\n", max(0, m.alturaCorpo()-lipgloss.Height(corpo))) + cortar(rodape, m.largura)
}

// alturaCorpo é o espaço entre o cabeçalho (2 linhas + 1 em branco) e o rodapé.
func (m modelo) alturaCorpo() int { return max(5, m.altura-5) }

func (m modelo) cabecalho() string {
	partes := []string{estiloMarca.Render("post-omarchy " + m.cfg.Versao), estiloSutil.Render(m.cfg.NomeSistema)}
	if m.conta.Nome != "" {
		partes = append(partes, estiloSutil.Render("usuário ")+estiloNegrito.Render(m.conta.Nome))
	}
	if m.cfg.DryRun {
		partes = append(partes, estiloSelo.Render("DRY-RUN"))
	}
	topo := strings.Join(partes, "  ")

	etapas := []string{"Usuário", "Módulos", "Revisão", "Execução", "Resumo"}
	atual := map[tela]int{telaUsuario: 0, telaOutroUsuario: 0, telaModulos: 1, telaRevisao: 2,
		telaExecucao: 3, telaResumo: 4}[m.tela]
	var trilha []string
	for i, e := range etapas {
		switch {
		case i == atual:
			trilha = append(trilha, estiloDestaque.Render(e))
		case i < atual:
			trilha = append(trilha, estiloNegrito.Render(e))
		default:
			trilha = append(trilha, estiloSutil.Render(e))
		}
	}
	return topo + "\n" + strings.Join(trilha, estiloSutil.Render(" › "))
}

func ajuda(pares ...string) string {
	var partes []string
	for i := 0; i+1 < len(pares); i += 2 {
		partes = append(partes, estiloNegrito.Render(pares[i])+" "+estiloSutil.Render(pares[i+1]))
	}
	return strings.Join(partes, estiloSutil.Render("  ·  "))
}

func cortar(texto string, largura int) string {
	return ansi.Truncate(texto, max(1, largura), "…")
}

func duracao(d time.Duration) string {
	s := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%dm%02ds", s/60, s%60)
}

// ------------------------------------------------------------------
// Usuário
// ------------------------------------------------------------------

func (m modelo) viewUsuario() string {
	var b strings.Builder
	b.WriteString(estiloNegrito.Render("Para qual usuário devo configurar o ambiente?") + "\n")
	b.WriteString(estiloSutil.Render("Shell, grupos, dots em ~/.config e apps do usuário serão aplicados a essa conta.") + "\n\n")

	for i := 0; i < m.totalContas(); i++ {
		var nome, detalhe string
		if i < len(m.contas) {
			c := m.contas[i]
			nome = c.Nome
			detalhe = c.Home
			if c.Gecos != "" {
				detalhe = c.Gecos + " · " + c.Home
			}
			if c.Nome == m.voce {
				detalhe += "  (você)"
			}
		} else {
			nome, detalhe = "Outro usuário…", "conta de domínio ou que não aparece na lista"
		}
		if i == m.cursorConta {
			b.WriteString(estiloDestaque.Render("› "+fmt.Sprintf("%-16s", nome)) + "  " + estiloSutil.Render(detalhe) + "\n")
		} else {
			b.WriteString("  " + fmt.Sprintf("%-16s", nome) + "  " + estiloSutil.Render(detalhe) + "\n")
		}
	}
	return b.String()
}

func (m modelo) viewOutroUsuario() string {
	var b strings.Builder
	b.WriteString(estiloNegrito.Render("Nome do usuário") + "\n")
	b.WriteString(estiloSutil.Render("Aceita contas locais e de domínio (AD/LDAP), resolvidas pelo getent.") + "\n\n")
	b.WriteString(m.entrada.View() + "\n")
	if m.erroEntrada != "" {
		b.WriteString("\n" + estiloErro.Render("✗ "+m.erroEntrada) + "\n")
	}
	return b.String()
}

// ------------------------------------------------------------------
// Módulos
// ------------------------------------------------------------------

func (m modelo) viewModulos() string {
	lista := m.cfg.Modulos
	titulo := estiloNegrito.Render("Escolha o que executar") + "  " +
		estiloSutil.Render(fmt.Sprintf("%d de %d selecionados", m.selecionados(), len(lista)))

	lado := m.largura >= 100
	larguraLista := m.largura
	if lado {
		larguraLista = m.largura * 55 / 100
	}
	altura := m.alturaCorpo() - 3 // título, linha em branco e nota
	detalhe := m.detalheModulo(m.cursor, m.largura-larguraLista-2, 0, false)
	if !lado {
		// Altura fixa (a do maior detalhe) para a lista não pular ao mover.
		maior := 0
		for i := range lista {
			maior = max(maior, lipgloss.Height(m.detalheModulo(i, m.largura-4, 0, true)))
		}
		detalhe = m.detalheModulo(m.cursor, m.largura-4, maior-2, true)
		altura -= maior
	}
	altura = max(3, altura)

	// Janela de rolagem que mantém o cursor visível.
	topo := max(0, min(m.cursor-altura/2, len(lista)-altura))
	var linhas []string
	for i := topo; i < min(len(lista), topo+altura); i++ {
		caixa := estiloSutil.Render("[ ]")
		if m.sel[i] {
			caixa = estiloDestaque.Render("[✔]")
		}
		marcador := "  "
		nome := lista[i].Titulo
		if i == m.cursor {
			marcador = estiloDestaque.Render("› ")
			nome = estiloNegrito.Render(nome)
		}
		linha := fmt.Sprintf("%s%s %s %s", marcador, caixa, estiloSutil.Render(fmt.Sprintf("%02d", i+1)), nome)
		linhas = append(linhas, cortar(linha, larguraLista-1))
	}
	if topo > 0 {
		linhas[0] = estiloSutil.Render(fmt.Sprintf("   ↑ mais %d", topo))
	}
	if resto := len(lista) - (topo + altura); resto > 0 {
		linhas[len(linhas)-1] = estiloSutil.Render(fmt.Sprintf("   ↓ mais %d", resto))
	}
	coluna := strings.Join(linhas, "\n")

	var corpo string
	if lado {
		corpo = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(larguraLista).Render(coluna), detalhe)
	} else {
		corpo = coluna + "\n" + detalhe
	}
	nota := ""
	if m.nota != "" {
		nota = estiloAviso.Render("! " + m.nota)
	}
	return titulo + "\n\n" + corpo + "\n" + nota
}

// detalheModulo: compacto corta cada campo em uma linha (terminal estreito,
// onde o detalhe divide a altura com a lista).
func (m modelo) detalheModulo(i, largura, altura int, compacto bool) string {
	lista := m.cfg.Modulos
	mod := lista[i]
	interna := max(20, largura-4)
	texto := lipgloss.NewStyle().Width(interna)
	if compacto {
		texto = lipgloss.NewStyle().MaxWidth(interna).Inline(true)
	}

	partes := []string{estiloNegrito.Render(mod.Titulo), texto.Render(mod.Descricao)}
	if deps := modulos.Titulos(lista, mod.Deps); len(deps) > 0 {
		partes = append(partes, texto.Render(estiloSutil.Render("Depende de: ")+strings.Join(deps, ", ")))
	}
	var precisam []string
	for _, outro := range lista {
		for _, d := range outro.Deps {
			if d == mod.ID {
				precisam = append(precisam, outro.Titulo)
			}
		}
	}
	if len(precisam) > 0 {
		partes = append(partes, texto.Render(estiloSutil.Render("Necessário para: ")+strings.Join(precisam, ", ")))
	}
	if mod.Interativo {
		partes = append(partes, texto.Render(estiloAviso.Render("⌨ Pode pedir sua senha ou uma confirmação.")))
	}
	partes = append(partes, estiloSutil.Render("id: "+mod.ID))
	return estiloCaixa.Width(largura - 2).Height(altura).Render(strings.Join(partes, "\n"))
}

// ------------------------------------------------------------------
// Revisão
// ------------------------------------------------------------------

// colunas distribui itens em quantas colunas forem precisas para caber em
// maxLinhas; se nem assim couber na largura, corta com "… e mais N".
func colunas(itens []string, maxLinhas, largura int) string {
	if len(itens) == 0 {
		return ""
	}
	maxLinhas = max(1, maxLinhas)
	larguraItem := 0
	for _, it := range itens {
		larguraItem = max(larguraItem, lipgloss.Width(it))
	}
	larguraItem = min(larguraItem+3, largura)
	// Colunas que caberiam com o título inteiro; se não bastarem, encolhe os
	// títulos (até 28) para caber mais colunas antes de cortar itens.
	precisa := (len(itens) + maxLinhas - 1) / maxLinhas
	nColunas := max(1, min(precisa, largura/larguraItem))
	if nColunas < precisa {
		nColunas = max(1, min(precisa, largura/min(larguraItem, 28)))
		larguraItem = min(larguraItem, largura/nColunas)
	}
	cabe := nColunas * maxLinhas
	if len(itens) > cabe {
		resto := len(itens) - cabe + 1
		itens = append(append([]string(nil), itens[:cabe-1]...), estiloSutil.Render(fmt.Sprintf("… e mais %d", resto)))
	}
	porColuna := (len(itens) + nColunas - 1) / nColunas
	var blocos []string
	for c := 0; c < nColunas; c++ {
		ini, fim := c*porColuna, min(len(itens), (c+1)*porColuna)
		if ini >= fim {
			break
		}
		var linhas []string
		for _, it := range itens[ini:fim] {
			linhas = append(linhas, cortar(it, larguraItem-1))
		}
		blocos = append(blocos, lipgloss.NewStyle().Width(larguraItem).Render(strings.Join(linhas, "\n")))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, blocos...)
}

func (m modelo) viewRevisao() string {
	var b strings.Builder
	b.WriteString(estiloNegrito.Render("Revise antes de começar") + "\n\n")

	modo := "Execução real — o sistema será alterado"
	if m.cfg.DryRun {
		modo = estiloDryRun.Render("Dry-run — só mostra o que seria feito")
	}
	campo := func(nome, valor string) {
		b.WriteString("  " + estiloSutil.Render(fmt.Sprintf("%-14s", nome)) + valor + "\n")
	}
	campo("Usuário-alvo", estiloNegrito.Render(m.conta.Nome)+estiloSutil.Render("  "+m.conta.Home))
	campo("Modo", modo)
	campo("Log", m.relativo(m.cfg.Log.Caminho))
	b.WriteString("\n")

	var itens []string
	interativos := false
	n := 0
	for i, mod := range m.cfg.Modulos {
		if !m.sel[i] {
			continue
		}
		n++
		item := fmt.Sprintf("%s %s", estiloSutil.Render(fmt.Sprintf("%2d.", n)), mod.Titulo)
		if mod.Interativo {
			item += " " + estiloAviso.Render("⌨")
			interativos = true
		}
		itens = append(itens, item)
	}
	b.WriteString(estiloNegrito.Render(fmt.Sprintf("  Módulos (%d)", n)) + "\n")

	var rodape []string
	if interativos {
		rodape = append(rodape, cortar(estiloAviso.Render("⌨")+estiloSutil.Render(
			" pode pedir sua senha (usa o terminal direto) ou uma confirmação."), m.largura))
	}
	for _, nota := range m.notas {
		rodape = append(rodape, cortar(estiloAviso.Render("! "+nota), m.largura))
	}
	ocupado := lipgloss.Height(b.String()) + len(rodape) + 1
	b.WriteString(indentar(colunas(itens, m.alturaCorpo()-ocupado, m.largura-4), "  ") + "\n")
	if len(rodape) > 0 {
		b.WriteString("\n" + strings.Join(rodape, "\n"))
	}
	return b.String()
}

// relativo encurta caminhos dentro do projeto (logs/...), que é de onde o
// usuário rodou o setup.
func (m modelo) relativo(caminho string) string {
	if rel, err := filepath.Rel(m.cfg.Raiz, caminho); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return caminho
}

func indentar(texto, prefixo string) string {
	linhas := strings.Split(texto, "\n")
	for i := range linhas {
		linhas[i] = prefixo + linhas[i]
	}
	return strings.Join(linhas, "\n")
}

// ------------------------------------------------------------------
// Execução
// ------------------------------------------------------------------

func (m modelo) icone(e Estado) string {
	switch e {
	case Rodando:
		return m.spinner.View()
	case Concluido:
		return estiloDestaque.Render("✓")
	case Falhou:
		return estiloErro.Render("✗")
	case NaoExecutado:
		return estiloSutil.Render("–")
	}
	return estiloSutil.Render("·")
}

func (m modelo) progresso() string {
	feitos := 0
	for _, e := range m.estados {
		if e == Concluido || e == Falhou {
			feitos++
		}
	}
	total := len(m.fila)
	decorrido := time.Since(m.inicio)
	if m.terminou {
		decorrido = m.duracao
	}
	contagem := fmt.Sprintf(" %d/%d  %s", feitos, total, duracao(decorrido))
	larguraBarra := max(10, min(40, m.largura-lipgloss.Width(contagem)-30))
	cheio := larguraBarra * feitos / max(1, total)
	cor := estiloDestaque
	if m.abortado || m.resultado().Falha != nil {
		cor = estiloErro
	}
	barra := cor.Render(strings.Repeat("━", cheio)) + estiloSutil.Render(strings.Repeat("━", larguraBarra-cheio))

	status := ""
	if !m.terminou && m.atual < total {
		status = "  " + estiloNegrito.Render(m.fila[m.atual].Titulo) +
			estiloSutil.Render(" · "+duracao(time.Since(m.inicioModulo)))
	}
	return cortar(barra+estiloSutil.Render(contagem)+status, m.largura)
}

func (m modelo) alturaSaida() int { return max(3, m.alturaCorpo()-4) }

func (m modelo) viewExecucao() string {
	altura := m.alturaSaida()
	larguraModulos := 0
	if m.largura >= 90 {
		larguraModulos = min(38, m.largura/3)
	}
	larguraSaida := m.largura - larguraModulos

	saida := m.viewSaida(larguraSaida-4, altura)
	tituloSaida := "Saída"
	if m.rolagem > 0 {
		tituloSaida += estiloAviso.Render(fmt.Sprintf("  (rolagem pausada, %d linhas abaixo — end para seguir)", m.rolagem))
	}
	painelSaida := estiloCaixa.Width(larguraSaida - 2).Height(altura).Render(saida)

	corpo := painelSaida
	if larguraModulos > 0 {
		var linhas []string
		topo := max(0, min(m.atual-altura/2, len(m.fila)-altura))
		for i := topo; i < min(len(m.fila), topo+altura); i++ {
			nome := m.fila[i].Titulo
			switch m.estados[i] {
			case Rodando:
				nome = estiloNegrito.Render(nome)
			case Pendente, NaoExecutado:
				nome = estiloSutil.Render(nome)
			}
			linhas = append(linhas, cortar(m.icone(m.estados[i])+" "+nome, larguraModulos-4))
		}
		painelModulos := estiloCaixa.Width(larguraModulos - 2).Height(altura).Render(strings.Join(linhas, "\n"))
		corpo = lipgloss.JoinHorizontal(lipgloss.Top, painelModulos, painelSaida)
	}
	return m.progresso() + "\n" + estiloSutil.Render(cortar(tituloSaida, m.largura)) + "\n" + corpo
}

// viewSaida desenha o fim da saída (ou o trecho rolado), quebrando linhas
// longas na largura do painel.
func (m modelo) viewSaida(largura, altura int) string {
	largura = max(10, largura)
	fim := len(m.linhas) - m.rolagem
	var visiveis []string
	for i := fim - 1; i >= 0 && len(visiveis) < altura; i-- {
		partes := strings.Split(estilizar(m.linhas[i], largura), "\n")
		for j := len(partes) - 1; j >= 0 && len(visiveis) < altura; j-- {
			visiveis = append(visiveis, partes[j])
		}
	}
	for i, j := 0, len(visiveis)-1; i < j; i, j = i+1, j-1 {
		visiveis[i], visiveis[j] = visiveis[j], visiveis[i]
	}
	return strings.Join(visiveis, "\n")
}

func estilizar(l linha, largura int) string {
	prefixo, estilo := "", lipgloss.NewStyle()
	switch l.tipo {
	case linhaInfo:
		prefixo, estilo = "• ", estiloInfo
	case linhaSucesso:
		prefixo, estilo = "✓ ", estiloDestaque
	case linhaAviso:
		prefixo, estilo = "⚠ ", estiloAviso
	case linhaErro:
		prefixo, estilo = "✗ ", estiloErro.Bold(true)
	case linhaDryRun:
		prefixo, estilo = "[dry-run] ", estiloDryRun
	case linhaPasso:
		prefixo, estilo = "▶ ", estiloPasso
	case linhaSaida:
		estilo = estiloSutil
	}
	// Continuações ficam alinhadas depois do prefixo.
	recuo := lipgloss.Width(prefixo)
	if recuo > 4 {
		recuo = 2
	}
	linhas := strings.Split(ansi.Wrap(l.texto, max(5, largura-recuo), ""), "\n")
	for i := range linhas {
		inicio := strings.Repeat(" ", recuo)
		if i == 0 {
			inicio = prefixo
		}
		linhas[i] = estilo.Render(inicio + linhas[i])
	}
	return strings.Join(linhas, "\n")
}

// ------------------------------------------------------------------
// Confirmação (plugins, temas)
// ------------------------------------------------------------------

func (m modelo) larguraModal() int { return min(78, m.largura-4) }

// corpoModal devolve o texto + itens da confirmação e quantas linhas cabem;
// o que não couber rola com ↑/↓ (o aviso de riscos não pode ser cortado).
func (m modelo) corpoModal() ([]string, int) {
	c := m.modal.c
	corpo := append([]string(nil), c.Texto...)
	for _, it := range c.Itens {
		corpo = append(corpo, estiloSutil.Render("  • ")+cortar(it, m.larguraModal()-10))
	}
	// Borda+padding (4), título+linha em branco (2), pergunta/botões/ajuda (6).
	return corpo, max(3, m.altura-12)
}

func (m modelo) viewModal() string {
	c := m.modal.c
	corTitulo := estiloNegrito
	borda := corAmarelo
	if c.Perigo {
		corTitulo, borda = estiloErro.Bold(true), corVermelho
	}

	corpo, cabe := m.corpoModal()
	titulo := corTitulo.Render(c.Titulo)
	teclas := []string{"←/→", "escolher", "enter", "confirmar", "s/n", "atalho"}
	if len(corpo) > cabe {
		topo := min(m.modalTopo, len(corpo)-cabe)
		titulo += estiloSutil.Render(fmt.Sprintf("  (%d–%d de %d)", topo+1, topo+cabe, len(corpo)))
		corpo = corpo[topo : topo+cabe]
		teclas = append([]string{"↑/↓", "rolar"}, teclas...)
	}

	botao := func(rotulo string, ativo bool, cor lipgloss.Color) string {
		estilo := lipgloss.NewStyle().Padding(0, 2)
		if ativo {
			return estilo.Bold(true).Foreground(lipgloss.Color("0")).Background(cor).Render(rotulo)
		}
		return estilo.Foreground(corCinza).Render(rotulo)
	}
	botoes := botao("Sim", m.modalSim, corVerde) + "  " + botao("Não", !m.modalSim, corVermelho)

	partes := append([]string{titulo, ""}, corpo...)
	partes = append(partes, "", estiloNegrito.Render(c.Pergunta), "", botoes, "", ajuda(teclas...))
	caixa := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borda).
		Padding(1, 2).Width(m.larguraModal()).Render(strings.Join(partes, "\n"))
	return lipgloss.Place(m.largura, m.altura, lipgloss.Center, lipgloss.Center, caixa)
}

// ------------------------------------------------------------------
// Resumo
// ------------------------------------------------------------------

func (m modelo) viewResumo() string {
	r := m.resultado()
	var b strings.Builder
	switch {
	case r.Abortado:
		b.WriteString(estiloErro.Bold(true).Render("■ Instalação abortada"))
	case r.Falha != nil:
		b.WriteString(estiloErro.Bold(true).Render("✗ A instalação parou num erro"))
	default:
		b.WriteString(estiloDestaque.Render("✓ Tudo pronto!"))
	}
	b.WriteString(estiloSutil.Render(fmt.Sprintf("  %d módulos · %s", len(r.Modulos), duracao(r.Duracao))) + "\n\n")

	larguraTitulo := 0
	for _, mod := range r.Modulos {
		larguraTitulo = max(larguraTitulo, lipgloss.Width(mod.Titulo))
	}
	larguraTitulo = min(larguraTitulo, 40)
	var itens []string
	for _, mod := range r.Modulos {
		titulo := cortar(mod.Titulo, larguraTitulo)
		item := m.icone(mod.Estado) + " " + titulo
		switch mod.Estado {
		case Concluido, Falhou:
			item += strings.Repeat(" ", larguraTitulo-lipgloss.Width(titulo)+2) + estiloSutil.Render(duracao(mod.Duracao))
		case NaoExecutado:
			item = estiloSutil.Render(m.icone(mod.Estado) + " " + titulo + "  (não executado)")
		}
		itens = append(itens, item)
	}

	var depois []string
	if r.Falha != nil && !r.Abortado {
		depois = append(depois, "", estiloErro.Render("Erro: ")+r.Falha.Error())
		depois = append(depois, estiloSutil.Render("Corrija e rode de novo: os módulos são idempotentes."))
	}
	if len(r.Avisos) > 0 {
		depois = append(depois, "", estiloAviso.Bold(true).Render(fmt.Sprintf("Avisos (%d)", len(r.Avisos))))
		limite := 6
		for i, a := range r.Avisos {
			if i == limite {
				depois = append(depois, estiloSutil.Render(fmt.Sprintf("  … e mais %d no log", len(r.Avisos)-limite)))
				break
			}
			depois = append(depois, cortar(estiloAviso.Render("  ⚠ ")+a, m.largura))
		}
	}
	depois = append(depois, "", estiloNegrito.Render("Próximos passos"))
	if r.Falha == nil && !r.Abortado && !m.cfg.DryRun {
		depois = append(depois, "  • Reinicie a máquina para aplicar as mudanças de grupo e kernel.")
	}
	depois = append(depois,
		"  • Log:            "+estiloSutil.Render(m.relativo(m.cfg.Log.Caminho)),
		"  • Saída completa: "+estiloSutil.Render(m.relativo(m.cfg.Log.CaminhoBruto)))

	ocupado := 2 + len(depois) + 1
	b.WriteString(indentar(colunas(itens, m.alturaCorpo()-ocupado, m.largura-2), " ") + "\n")
	b.WriteString(strings.Join(depois, "\n"))
	return b.String()
}
