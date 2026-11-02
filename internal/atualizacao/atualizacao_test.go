package atualizacao

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMaisNova(t *testing.T) {
	casos := []struct {
		a, b     string
		esperado bool
	}{
		{"v1.6.0", "v1.5.0", true},
		{"v1.5.1", "v1.5.0", true},
		{"v2.0.0", "v1.99.99", true},
		{"v1.10.0", "v1.9.0", true},
		{"v1.5.0", "v1.5.0", false},
		{"v1.4.9", "v1.5.0", false},
		{"v1.6.0", "dev", false},
		{"lixo", "v1.5.0", false},
	}
	for _, c := range casos {
		if obtido := MaisNova(c.a, c.b); obtido != c.esperado {
			t.Errorf("MaisNova(%q, %q) = %v, esperado %v", c.a, c.b, obtido, c.esperado)
		}
	}
}

const raizTeste = "post-omarchy-v9.9.9/"

var pacoteTeste = map[string]string{
	raizTeste + "post-omarchy":       "binário novo",
	raizTeste + "setup.sh":           "setup novo",
	raizTeste + "README.md":          "readme novo",
	raizTeste + "OVPN/README.md":     "ovpn readme novo",
	raizTeste + "data/packages.json": "dados novos",
	raizTeste + "dots/hypr/novo.lua": "dot novo",
	raizTeste + "assets/icone.svg":   "ícone novo",
	// Nada disto pode ser escrito:
	raizTeste + "OVPN/empresa.ovpn": "ovpn falso",
	raizTeste + "logs/x.log":        "log falso",
	raizTeste + "../fora.txt":       "fuga",
	"/etc/cron.d/x":                 "absoluto",
	"outra-pasta/post-omarchy":      "fora da raiz",
}

func montarPacote(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for nome, conteudo := range pacoteTeste {
		w, err := z.Create(nome)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(conteudo)); err != nil {
			t.Fatal(err)
		}
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestExtrair(t *testing.T) {
	pasta := t.TempDir()
	pacote := filepath.Join(pasta, "p.zip")
	if err := os.WriteFile(pacote, montarPacote(t), 0o644); err != nil {
		t.Fatal(err)
	}

	obtidos, err := extrair(pacote, filepath.Join(pasta, "x"), raizTeste, true)
	if err != nil {
		t.Fatal(err)
	}
	esperados := []string{"post-omarchy", "setup.sh", "README.md", "OVPN/README.md",
		"data/packages.json", "dots/hypr/novo.lua", "assets/icone.svg"}
	if len(obtidos) != len(esperados) {
		t.Errorf("extrair = %v, esperado só %v", obtidos, esperados)
	}
	for _, e := range esperados {
		if _, ok := obtidos[e]; !ok {
			t.Errorf("faltou %s", e)
		}
	}

	// Binário solto, fora do pacote: só o binário.
	obtidos, err = extrair(pacote, filepath.Join(pasta, "y"), raizTeste, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := obtidos["post-omarchy"]; len(obtidos) != 1 || !ok {
		t.Errorf("extrair sem pacote completo = %v", obtidos)
	}
}

func TestAplicar(t *testing.T) {
	pacote := montarPacote(t)
	soma := sha256.Sum256(pacote)
	hash := hex.EncodeToString(soma[:])

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pacote":
			w.Write(pacote)
		case "/sha256":
			w.Write([]byte(hash + "  pacote.zip\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	antigo := cliente
	cliente = srv.Client()
	defer func() { cliente = antigo }()

	base := t.TempDir()
	for nome, conteudo := range map[string]string{
		"post-omarchy":          "binário velho",
		"setup.sh":              "setup velho",
		"data/packages.json":    "dados velhos",
		"dots/claude/CLAUDE.md": "minha memória",
		"OVPN/empresa.ovpn":     "meu perfil",
		"logs/setup.log":        "meu log",
	} {
		caminho := filepath.Join(base, nome)
		os.MkdirAll(filepath.Dir(caminho), 0o755)
		if err := os.WriteFile(caminho, []byte(conteudo), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := &Release{Tag: "v9.9.9", url: srv.URL + "/pacote", sumURL: srv.URL + "/sha256", digest: hash}
	exe := filepath.Join(base, "post-omarchy")
	if err := Aplicar(context.Background(), r, exe); err != nil {
		t.Fatal(err)
	}
	for nome, esperado := range map[string]string{
		"post-omarchy":          "binário novo",
		"setup.sh":              "setup novo",
		"data/packages.json":    "dados novos",
		"dots/hypr/novo.lua":    "dot novo",
		"dots/claude/CLAUDE.md": "minha memória",
		"OVPN/empresa.ovpn":     "meu perfil",
		"logs/setup.log":        "meu log",
	} {
		obtido, err := os.ReadFile(filepath.Join(base, nome))
		if err != nil || string(obtido) != esperado {
			t.Errorf("%s = %q (%v), esperado %q", nome, obtido, err, esperado)
		}
	}
	for _, nome := range []string{"post-omarchy", "setup.sh"} {
		if info, _ := os.Stat(filepath.Join(base, nome)); info.Mode().Perm() != 0o755 {
			t.Errorf("%s com permissão %o", nome, info.Mode().Perm())
		}
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(base), "fora.txt")); err == nil {
		t.Error("entrada com ../ escreveu fora da pasta")
	}
	if sobras, _ := filepath.Glob(filepath.Join(base, ".post-omarchy-update-*")); len(sobras) > 0 {
		t.Errorf("diretório temporário não removido: %v", sobras)
	}

	// SHA-256 que não confere: nada é alterado.
	r.digest = "00" + hash[2:]
	os.WriteFile(exe, []byte("binário velho"), 0o755)
	if err := Aplicar(context.Background(), r, exe); err == nil {
		t.Fatal("Aplicar aceitou um pacote com SHA-256 diferente")
	}
	if obtido, _ := os.ReadFile(exe); string(obtido) != "binário velho" {
		t.Errorf("binário alterado mesmo com SHA-256 errado: %q", obtido)
	}
}

// No clone do repositório (com main.go), o setup.sh recompila: só o binário muda.
func TestAplicarNoClone(t *testing.T) {
	pacote := montarPacote(t)
	soma := sha256.Sum256(pacote)
	hash := hex.EncodeToString(soma[:])
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pacote" {
			w.Write(pacote)
			return
		}
		w.Write([]byte(hash + "  pacote.zip\n"))
	}))
	defer srv.Close()
	antigo := cliente
	cliente = srv.Client()
	defer func() { cliente = antigo }()

	base := t.TempDir()
	os.MkdirAll(filepath.Join(base, "data"), 0o755)
	for nome, conteudo := range map[string]string{"main.go": "código", "post-omarchy": "velho", "data/packages.json": "meus dados"} {
		if err := os.WriteFile(filepath.Join(base, nome), []byte(conteudo), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := &Release{Tag: "v9.9.9", url: srv.URL + "/pacote", sumURL: srv.URL + "/sha256"}
	if err := Aplicar(context.Background(), r, filepath.Join(base, "post-omarchy")); err != nil {
		t.Fatal(err)
	}
	if obtido, _ := os.ReadFile(filepath.Join(base, "data/packages.json")); string(obtido) != "meus dados" {
		t.Errorf("clone teve data/ alterado: %q", obtido)
	}
}
