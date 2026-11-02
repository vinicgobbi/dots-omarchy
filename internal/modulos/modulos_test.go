package modulos

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

// uiTeste guarda o que os módulos mostrariam.
type uiTeste struct{ dryRun, avisos []string }

func (u *uiTeste) Info(string)                   {}
func (u *uiTeste) Sucesso(string)                {}
func (u *uiTeste) Aviso(msg string)              { u.avisos = append(u.avisos, msg) }
func (u *uiTeste) DryRun(texto string)           { u.dryRun = append(u.dryRun, texto) }
func (u *uiTeste) Saida() io.Writer              { return io.Discard }
func (u *uiTeste) Interativo() bool              { return false }
func (u *uiTeste) Confirmar(ui.Confirmacao) bool { return false }
func (u *uiTeste) Isolado() bool                 { return false }

func modulosFalsos(erros map[string]error) []*Modulo {
	novo := func(id string, deps ...string) *Modulo {
		return &Modulo{ID: id, Titulo: id, Deps: deps, Executar: func(*sistema.Sistema) error { return erros[id] }}
	}
	return []*Modulo{novo("a"), novo("b", "a"), novo("c", "b"), novo("d")}
}

func TestIDsUnicos(t *testing.T) {
	vistos := map[string]bool{}
	for _, m := range Todos() {
		if vistos[m.ID] {
			t.Errorf("ID duplicado: %s", m.ID)
		}
		vistos[m.ID] = true
		if m.Executar == nil {
			t.Errorf("%s sem função Executar", m.ID)
		}
	}
}

func TestGruposContiguos(t *testing.T) {
	ordem := []string{GrupoSistema, GrupoTerminal, GrupoAplicativos, GrupoDesenvolvimento, GrupoRede, GrupoOmarchy, GrupoFinalizacao}
	posicao := -1
	for _, m := range Todos() {
		g := slices.Index(ordem, m.Grupo)
		if g < 0 {
			t.Errorf("%s: grupo desconhecido %q", m.ID, m.Grupo)
			continue
		}
		if g < posicao {
			t.Errorf("%s: grupo %q fora de ordem (os módulos de cada grupo precisam ficar juntos)", m.ID, m.Grupo)
		}
		posicao = g
	}
}

func TestDependenciasRodamAntes(t *testing.T) {
	todos := Todos()
	for i, m := range todos {
		for _, d := range m.Deps {
			j := indice(todos, d)
			if j < 0 {
				t.Errorf("%s depende de %q, que não existe", m.ID, d)
			} else if j >= i {
				t.Errorf("%s depende de %s, que roda depois dele", m.ID, d)
			}
		}
	}
}

// O shell troca o ~/.zshrc, e a limpeza precisa vir depois de tudo.
func TestOrdemDeExecucao(t *testing.T) {
	todos := Todos()
	for _, id := range []string{"node", "ruby"} {
		if indice(todos, "shell") > indice(todos, id) {
			t.Errorf("shell precisa rodar antes de %s", id)
		}
	}
	if todos[len(todos)-1].ID != "limpeza" {
		t.Errorf("o último módulo é %s, esperado limpeza", todos[len(todos)-1].ID)
	}
}

func TestMarcarEmCadeia(t *testing.T) {
	fila := modulosFalsos(nil)
	sel := make([]bool, len(fila))
	puxados := Marcar(fila, sel, 2)
	if !slices.Equal(sel, []bool{true, true, true, false}) || len(puxados) != 2 {
		t.Errorf("sel = %v, puxados = %v", sel, puxados)
	}
}

func TestConexoesNM(t *testing.T) {
	obtido := conexoesNM("Casa:802-11-wireless\nEmpresa\\: VPN:vpn\n\nsem-tipo\n")
	if len(obtido) != 2 || obtido["Casa"] != "802-11-wireless" || obtido["Empresa: VPN"] != "vpn" {
		t.Errorf("conexoesNM = %v", obtido)
	}
}

// Conexão homônima que não é VPN (Wi-Fi, cabeada) não pode ser apagada.
func TestOVPNNaoApagaConexaoQueNaoEVPN(t *testing.T) {
	pasta := t.TempDir()
	if err := os.Mkdir(filepath.Join(pasta, "OVPN"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, nome := range []string{"casa.ovpn", "vpn-antiga.ovpn"} {
		if err := os.WriteFile(filepath.Join(pasta, "OVPN", nome), []byte("client\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(pasta, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	falso := "#!/bin/sh\nprintf 'casa:802-11-wireless\\nvpn-antiga:vpn\\n'\n"
	if err := os.WriteFile(filepath.Join(bin, "nmcli"), []byte(falso), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "pacman"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	d, err := dados.Carregar("../..")
	if err != nil {
		t.Fatal(err)
	}
	u := &uiTeste{}
	s := sistema.Novo(context.Background(), pasta, sistema.Conta{}, d, true, u, nil)
	if err := importarOVPN(s); err != nil {
		t.Fatal(err)
	}
	comandos := strings.Join(u.dryRun, "\n")
	if strings.Contains(comandos, "delete id casa") {
		t.Errorf("apagou a conexão Wi-Fi homônima:\n%s", comandos)
	}
	if !strings.Contains(comandos, "delete id vpn-antiga") {
		t.Errorf("não reimportou a VPN existente:\n%s", comandos)
	}
	avisou := false
	for _, a := range u.avisos {
		avisou = avisou || strings.Contains(a, "qualquer usuário")
	}
	if !avisou {
		t.Errorf("não avisou sobre .ovpn legível por todos: %v", u.avisos)
	}
}

// Os scripts embutidos precisam ao menos ser shell válido.
func TestScriptsSaoShellValido(t *testing.T) {
	scripts := map[string]string{
		"autostart do Solaar": scriptAutostartSolaar,
		"extensões Nautilus":  scriptExtensoesNautilus(),
		"ruby":                scriptRuby([]string{"bundler", "jekyll"}),
	}
	for nome, script := range scripts {
		cmd := exec.Command("bash", "-n")
		cmd.Stdin = strings.NewReader(script)
		if saida, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s: %v\n%s", nome, err, saida)
		}
	}
}

func TestExtensoesNautilusFixadasESemTmpFixo(t *testing.T) {
	script := scriptExtensoesNautilus()
	for _, commit := range []string{vscodeNautilusCommit, copypathCommit, abrirTerminalCommit} {
		if !strings.Contains(script, commit) {
			t.Errorf("commit %s não aparece no script", commit)
		}
	}
	if strings.Contains(script, "/tmp/") {
		t.Errorf("caminho fixo em /tmp no script:\n%s", script)
	}
}

func TestRegraSolaarEmbutida(t *testing.T) {
	if !strings.Contains(regraSolaar, "hidraw") {
		t.Errorf("regra UDEV embutida inesperada: %q", regraSolaar)
	}
}

// Todo módulo roda de ponta a ponta no --dry-run, com os dados reais, sem
// alterar nada nem falhar por dado faltando.
func TestTodosOsModulosNoDryRun(t *testing.T) {
	d, err := dados.Carregar("../..")
	if err != nil {
		t.Fatal(err)
	}
	atual, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	uid, _ := strconv.Atoi(atual.Uid)
	gid, _ := strconv.Atoi(atual.Gid)
	raiz, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	conta := sistema.Conta{Nome: atual.Username, UID: uid, GID: gid, Home: t.TempDir()}

	for _, m := range Todos() {
		t.Run(m.ID, func(t *testing.T) {
			if _, err := os.Stat(pacmanConf); m.ID == "repositorios" && err != nil {
				t.Skip("sem " + pacmanConf + " nesta máquina")
			}
			u := &uiTeste{}
			s := sistema.Novo(context.Background(), raiz, conta, d, true, u, nil)
			s.Trabalho = t.TempDir()
			if err := m.Executar(s); err != nil {
				t.Fatal(err)
			}
			comandos := strings.Join(u.dryRun, "\n")
			// Plugins e temas não instalam sem terminal; o Claude Code pode já
			// estar instalado na máquina que roda o teste.
			semComandos := m.ID == "plugins_omarchy" || m.ID == "temas_omarchy" || m.ID == "claude_code"
			if !semComandos && comandos == "" {
				t.Error("o módulo não fez nada")
			}
			for _, proibido := range []string{"/tmp/gcm", "/tmp/dotfiles", "master/rules.d-uinput"} {
				if strings.Contains(comandos, proibido) {
					t.Errorf("comando com %q:\n%s", proibido, comandos)
				}
			}
		})
	}
}
