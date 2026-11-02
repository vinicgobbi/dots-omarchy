// post-omarchy: pós-instalação do Omarchy com menu de módulos, dependências,
// execução e log.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/x/term"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/logs"
	"github.com/vinicgobbi/dots-omarchy/internal/modulos"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/tui"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// versao é preenchida no build do release (-ldflags "-X main.versao=...").
var versao = "dev"

type opcoes struct {
	dryRun   bool
	modulos  string
	usuario  string
	semTUI   bool
	failFast bool
	listar   bool
	versao   bool
	raiz     string
}

func argumentos() opcoes {
	var o opcoes
	flag.BoolVar(&o.dryRun, "dry-run", false, "mostra os comandos e arquivos que seriam alterados, sem executar")
	flag.StringVar(&o.modulos, "modulos", "", "roda só estes módulos (ids separados por vírgula) e as dependências, sem o menu")
	flag.StringVar(&o.usuario, "usuario", "", "usuário-alvo (sem perguntar)")
	flag.BoolVar(&o.semTUI, "sem-tui", false, "usa a saída em texto corrido em vez da interface interativa")
	flag.BoolVar(&o.failFast, "fail-fast", false, "para tudo no primeiro módulo que falhar")
	flag.BoolVar(&o.listar, "listar", false, "lista os módulos e seus ids e sai")
	flag.BoolVar(&o.versao, "versao", false, "mostra a versão e sai")
	flag.StringVar(&o.raiz, "raiz", "", "pasta do projeto (com data/ e dots/); padrão: a do executável")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Pós-instalação do Omarchy.\n\nUso: sudo ./setup.sh [opções]\n\nOpções:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	return o
}

func main() {
	os.Exit(rodar())
}

func falhar(msg string) int {
	fmt.Fprintf(os.Stderr, "%s[-] %s%s\n", ui.Vermelho, msg, ui.NC)
	return 1
}

// acharRaiz procura a pasta com data/: a do executável (setup.sh compila
// para a raiz do projeto) ou, no `go run`, a atual.
func acharRaiz(flagRaiz string) (string, error) {
	candidatos := []string{flagRaiz}
	if flagRaiz == "" {
		if exe, err := os.Executable(); err == nil {
			if real, err := filepath.EvalSymlinks(exe); err == nil {
				candidatos = append(candidatos, filepath.Dir(real))
			}
		}
		if wd, err := os.Getwd(); err == nil {
			candidatos = append(candidatos, wd)
		}
	}
	for _, c := range candidatos {
		if c == "" {
			continue
		}
		if info, err := os.Stat(filepath.Join(c, "data")); err == nil && info.IsDir() {
			return filepath.Abs(c)
		}
	}
	return "", errors.New("pasta do projeto (com data/) não encontrada; use --raiz")
}

// detectarSistema: este projeto é exclusivo do Omarchy (Arch por baixo:
// pacman, yay pré-instalado, Hyprland).
func detectarSistema() (string, error) {
	conteudo, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	release := map[string]string{}
	for _, linha := range strings.Split(string(conteudo), "\n") {
		chave, valor, _ := strings.Cut(linha, "=")
		release[chave] = strings.Trim(valor, `"`)
	}
	nome := release["PRETTY_NAME"]
	if nome == "" {
		nome = "desconhecido"
	}
	if release["ID"] != "omarchy" {
		return nome, fmt.Errorf("sistema não suportado: %s; este programa é exclusivo do Omarchy (ID=omarchy) "+
			"(para o Ubuntu com GNOME, use o projeto post-install)", nome)
	}
	return nome, nil
}

func terminal(f *os.File) bool { return term.IsTerminal(f.Fd()) }

func rodar() int {
	o := argumentos()
	lista := modulos.Todos()

	if o.versao {
		fmt.Println("post-omarchy", versao)
		return 0
	}
	if o.listar {
		for i, m := range lista {
			if i == 0 || m.Grupo != lista[i-1].Grupo {
				if i > 0 {
					fmt.Println()
				}
				fmt.Println(m.Grupo + ":")
			}
			deps := ""
			if len(m.Deps) > 0 {
				deps = " (depende de: " + strings.Join(m.Deps, ", ") + ")"
			}
			fmt.Printf("  %02d  %-20s %s%s\n", i+1, m.ID, m.Titulo, deps)
		}
		return 0
	}

	if os.Geteuid() != 0 && !o.dryRun {
		return falhar("Execute como root (sudo ./setup.sh), ou use --dry-run para só simular.")
	}

	// Legível por todos: os scripts como usuário-alvo leem os downloads daqui.
	trabalho, err := os.MkdirTemp("", "post-omarchy-")
	if err != nil {
		return falhar(err.Error())
	}
	defer os.RemoveAll(trabalho)
	if err := os.Chmod(trabalho, 0o755); err != nil {
		return falhar(err.Error())
	}

	// A regra temporária de sudo (sistema.LiberarSudo) nunca pode sobrar: sai
	// na saída normal, em SIGTERM/SIGHUP (terminal fechado) e, se uma execução
	// anterior morreu com SIGKILL, aqui no início.
	sistema.RemoverSudo()
	defer sistema.RemoverSudo()
	sinais := make(chan os.Signal, 1)
	signal.Notify(sinais, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sinais
		sistema.RemoverSudo()
		os.RemoveAll(trabalho)
		os.Exit(143)
	}()
	raiz, err := acharRaiz(o.raiz)
	if err != nil {
		return falhar(err.Error())
	}
	nomeSistema, err := detectarSistema()
	if err != nil {
		return falhar(err.Error())
	}
	d, err := dados.Carregar(raiz)
	if err != nil {
		return falhar("Dados inválidos: " + err.Error())
	}

	var selecao []bool
	if o.modulos != "" {
		if selecao, err = modulos.Selecionar(lista, strings.Split(o.modulos, ",")); err != nil {
			return falhar(err.Error() + " (veja --listar)")
		}
	}

	var conta *sistema.Conta
	if o.usuario != "" {
		c, err := sistema.BuscarConta(o.usuario)
		if err != nil {
			return falhar(err.Error())
		}
		conta = &c
	}

	log, err := logs.Iniciar(raiz, "setup")
	if err != nil {
		return falhar("Não foi possível criar o log: " + err.Error())
	}
	defer log.Fechar()
	log.Linha("INFO", "Sistema detectado: "+nomeSistema)

	if terminal(os.Stdin) && terminal(os.Stdout) && !o.semTUI {
		cfg := tui.Config{
			Versao: versao, Raiz: raiz, Dados: d, DryRun: o.dryRun, Log: log, NomeSistema: nomeSistema,
			Modulos: lista, Selecao: selecao, PularMenu: selecao != nil, Conta: conta,
			FalhaRapida: o.failFast, Trabalho: trabalho,
		}
		return rodarTUI(cfg)
	}
	return rodarTexto(o, raiz, nomeSistema, d, lista, selecao, conta, log, trabalho)
}

// codigoSaida: 130 se algo foi cancelado, 1 se algo falhou.
func codigoSaida(resultados []modulos.Resultado) int {
	codigo := 0
	for _, r := range resultados {
		switch r.Estado {
		case modulos.Cancelado:
			return 130
		case modulos.Falhou:
			codigo = 1
		}
	}
	return codigo
}

func rodarTUI(cfg tui.Config) int {
	r, err := tui.Rodar(cfg)
	if err != nil {
		return falhar("Erro na interface: " + err.Error())
	}
	if !r.Iniciado {
		fmt.Println("Nada foi executado.")
		return 0
	}

	// A TUI some ao sair; deixa o resumo no terminal.
	fmt.Println(tui.Resumo(r, cfg.DryRun, cfg.Log.Caminho, cfg.Log.CaminhoBruto))
	return codigoSaida(r.Modulos)
}

// rodarTexto é o fluxo sem TUI: sem terminal interativo (ou com --sem-tui),
// roda todos os módulos ou os de --modulos.
func rodarTexto(o opcoes, raiz, nomeSistema string, d *dados.Dados, lista []*modulos.Modulo,
	selecao []bool, conta *sistema.Conta, log *logs.Log, trabalho string) int {
	interativo := terminal(os.Stdin)
	t := ui.NovoTexto(log, interativo)
	t.Banner()
	t.Info("Log desta execução: " + log.Caminho)
	t.Info("Sistema detectado: " + nomeSistema + " (Omarchy — Arch por baixo, pacman + Chaotic-AUR/yay)")

	if detectado := sistema.DetectarUsuario(); conta == nil && detectado != "" && interativo &&
		t.SimNaoPadraoSim(fmt.Sprintf("Configurar o ambiente para o usuário '%s'?", detectado)) {
		if c, err := sistema.BuscarConta(detectado); err == nil {
			conta = &c
		}
	}
	for conta == nil {
		nome, err := t.Pergunta("Para qual usuário do sistema devo configurar o ambiente? ")
		if err != nil {
			t.Erro("Sem resposta para o usuário-alvo; use --usuario.")
			return 1
		}
		if nome == "" {
			t.Aviso("Digite um nome de usuário.")
			continue
		}
		c, err := sistema.BuscarConta(nome)
		if err != nil {
			t.Aviso(err.Error() + ". Tente novamente.")
			continue
		}
		conta = &c
	}
	t.Info("Configuração será aplicada para: " + conta.Nome)
	if o.dryRun {
		t.Aviso("Modo --dry-run: nada será alterado no sistema.")
	}

	if selecao == nil {
		t.Aviso("Sem a interface interativa: executando todos os módulos (use --modulos para escolher).")
		selecao = make([]bool, len(lista))
		for i := range selecao {
			selecao[i] = true
		}
	}
	for _, p := range modulos.Resolver(lista, selecao) {
		t.Aviso(fmt.Sprintf("'%s' depende de '%s': selecionando automaticamente.", lista[p.Por].Titulo, lista[p.Dep].Titulo))
	}
	var escolhidos []*modulos.Modulo
	for i, m := range lista {
		if selecao[i] {
			escolhidos = append(escolhidos, m)
		}
	}
	if len(escolhidos) == 0 {
		t.Aviso("Nenhum módulo selecionado. Nada a fazer.")
		return 0
	}

	fmt.Println()
	t.Info("Módulos selecionados:")
	for _, m := range escolhidos {
		fmt.Printf("  - %s\n", m.Titulo)
	}
	fmt.Println()
	if interativo && !t.SimNao(fmt.Sprintf("Iniciar a configuração com os %d módulos acima?", len(escolhidos))) {
		t.Aviso("Operação cancelada pelo usuário.")
		return 0
	}

	if err := log.IniciarBruto(); err == nil {
		t.Info("Saída completa (bruta) desta execução: " + log.CaminhoBruto)
	}
	if !o.dryRun {
		if err := log.Chown(conta.UID, conta.GID); err != nil {
			t.Aviso(fmt.Sprintf("Não foi possível passar os logs para %s: %v", conta.Nome, err))
		}
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	s := sistema.Novo(ctx, raiz, *conta, d, o.dryRun, t, log)
	s.Trabalho = trabalho
	restaurar, err := s.LiberarSudo()
	if err != nil {
		t.Erro(err.Error())
		return 1
	}
	defer restaurar()

	inicio := time.Now()
	resultados := modulos.Rodar(ctx, s, escolhidos, o.failFast, modulos.Observador{
		Inicio: func(n int) { t.Passo(n+1, len(escolhidos), escolhidos[n].Titulo) },
		Fim: func(_ int, r modulos.Resultado) {
			if r.Estado == modulos.Falhou {
				t.Erro(fmt.Sprintf("'%s' falhou: %v", r.Modulo.Titulo, r.Erro))
			}
		},
	})

	fmt.Println()
	fmt.Println(tui.Resumo(tui.Resultado{Modulos: resultados, Duracao: time.Since(inicio), Abortado: ctx.Err() != nil},
		o.dryRun, log.Caminho, log.CaminhoBruto))
	return codigoSaida(resultados)
}
