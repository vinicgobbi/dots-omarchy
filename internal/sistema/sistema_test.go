package sistema

import (
	"context"
	"strings"
	"testing"

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
