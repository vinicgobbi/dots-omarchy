package sistema

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

func TestJuntar(t *testing.T) {
	casos := map[string][]string{
		"pacman -S --needed git":                 {"pacman", "-S", "--needed", "git"},
		"echo 'a b' 'it'\"'\"'s'":                {"echo", "a b", "it's"},
		"omarchy-webapp-install X https://x.y/z": {"omarchy-webapp-install", "X", "https://x.y/z"},
	}
	for esperado, args := range casos {
		if obtido := Juntar(args...); obtido != esperado {
			t.Errorf("Juntar(%q) = %q, esperado %q", args, obtido, esperado)
		}
	}
}

func TestEspelharDevolveStdoutEStderr(t *testing.T) {
	s := Novo(context.Background(), t.TempDir(), Conta{}, nil, false, ui.NovoTexto(nil, false), nil)
	res, err := s.Rodar([]string{"bash", "-c", "echo saida; echo 'plugin id x is already used by y' >&2; exit 1"},
		Opts{SemCheck: true, Espelhar: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Codigo != 1 {
		t.Errorf("código = %d, esperado 1", res.Codigo)
	}
	if !strings.Contains(res.Saida, "saida") || !strings.Contains(res.Saida, "is already used by") {
		t.Errorf("saída capturada incompleta: %q", res.Saida)
	}
}

func TestFalhaComCheck(t *testing.T) {
	s := Novo(context.Background(), t.TempDir(), Conta{}, nil, false, ui.NovoTexto(nil, false), nil)
	_, err := s.Rodar([]string{"false"}, Opts{Quieto: true})
	if f, ok := err.(*FalhaComando); !ok || f.Codigo != 1 {
		t.Errorf("esperava FalhaComando com código 1, veio %v", err)
	}
}

func TestVerificarSHA256(t *testing.T) {
	arquivo := filepath.Join(t.TempDir(), "a")
	if err := os.WriteFile(arquivo, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	const certo = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if err := VerificarSHA256(arquivo, strings.ToUpper(certo)); err != nil {
		t.Errorf("hash certo recusado: %v", err)
	}
	if err := VerificarSHA256(arquivo, "00"+certo[2:]); err == nil {
		t.Error("hash errado aceito")
	}
}

func TestSemDowngrade(t *testing.T) {
	seguro, _ := http.NewRequest(http.MethodGet, "https://a.b/x", nil)
	inseguro, _ := http.NewRequest(http.MethodGet, "http://a.b/x", nil)
	via := []*http.Request{seguro}
	if err := semDowngrade(seguro, via); err != nil {
		t.Errorf("https para https recusado: %v", err)
	}
	if err := semDowngrade(inseguro, via); err == nil {
		t.Error("https para http aceito")
	}
}

func TestGetFalhaEmErroHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "página de erro", http.StatusNotFound)
	}))
	defer srv.Close()
	s := Novo(context.Background(), t.TempDir(), Conta{}, nil, false, ui.NovoTexto(nil, false), nil)
	if _, err := s.Get(srv.URL, time.Second); err == nil {
		t.Error("Get aceitou um 404")
	}
	destino := filepath.Join(t.TempDir(), "x")
	if err := s.baixarUmaVez(srv.URL, destino); err == nil {
		t.Error("o download aceitou um 404")
	}
	if _, err := os.Stat(destino); err == nil {
		t.Error("a página de erro foi gravada no destino")
	}
}

func TestDetectarUsuarioIgnoraRoot(t *testing.T) {
	atual, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	if atual.Username == "root" {
		t.Skip("precisa de um usuário comum")
	}
	t.Setenv("SUDO_USER", "root")
	t.Setenv("PKEXEC_UID", "")
	t.Setenv("USER", atual.Username)
	if obtido := DetectarUsuario(); obtido == "root" {
		t.Errorf("DetectarUsuario = %q", obtido)
	}
	t.Setenv("SUDO_USER", atual.Username)
	if obtido := DetectarUsuario(); obtido != atual.Username {
		t.Errorf("DetectarUsuario = %q, esperado %q", obtido, atual.Username)
	}
}
