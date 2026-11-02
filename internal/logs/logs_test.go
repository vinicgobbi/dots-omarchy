package logs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLogsSoLegiveisPeloDono(t *testing.T) {
	antigo := unixUmask(0)
	defer unixUmask(antigo)

	l, err := Iniciar(t.TempDir(), "teste")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Fechar()
	if err := l.IniciarBruto(); err != nil {
		t.Fatal(err)
	}
	for _, caminho := range []string{l.Caminho, l.CaminhoBruto} {
		info, err := os.Stat(caminho)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s com permissão %o, esperado 600", filepath.Base(caminho), info.Mode().Perm())
		}
	}
}
