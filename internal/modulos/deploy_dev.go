package modulos

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

const deployDevRepo = "vinicgobbi/deploy-tools"

// O deploy-dev se instala sozinho (install copia para ~/.local/bin e põe a
// pasta no PATH) e depois se atualiza sozinho a partir das releases.
var deployDev = &Modulo{
	ID:        "deploy_dev",
	Grupo:     GrupoDesenvolvimento,
	Titulo:    "deploy-dev",
	Descricao: "CLI do deploy-tools que gera e valida o workflow de deploy dos projetos (github.com/vinicgobbi/deploy-tools)",
	Executar:  instalarDeployDev,
}

func instalarDeployDev(s *sistema.Sistema) error {
	s.UI.Info("Instalando o deploy-dev (última release do deploy-tools)...")
	binario, err := ultimoAsset(s, deployDevRepo, func(n string) bool { return n == "deploy-dev-linux-amd64" })
	if err != nil {
		return err
	}
	soma, err := ultimoAsset(s, deployDevRepo, func(n string) bool { return n == "deploy-dev-linux-amd64.sha256" })
	if err != nil {
		return err
	}
	destino := s.Temp("deploy-dev-linux-amd64")
	if err := baixarAsset(s, binario, destino); err != nil {
		return err
	}
	if !s.DryRun {
		if err := conferirSHA256Publicado(s, soma.URL, destino); err != nil {
			return err
		}
	}
	// O binário baixado é do root; o usuário copia para um temporário dele para
	// poder marcar como executável.
	if err := s.Usr(inicioScriptTemp + "install -m 0755 " + sistema.Juntar(destino) + " \"$dir/deploy-dev\"\n" +
		"\"$dir/deploy-dev\" install --add-to-path\n"); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("deploy-dev instalado em ~/.local/bin para %s.", s.Usuario()))
	return nil
}

// conferirSHA256Publicado compara o arquivo com o .sha256 publicado na release
// (formato do sha256sum: "<hash>  <nome>").
func conferirSHA256Publicado(s *sistema.Sistema, url, arquivo string) error {
	corpo, err := s.Get(url, 30*time.Second)
	if err != nil {
		return err
	}
	campos := strings.Fields(string(corpo))
	if len(campos) == 0 {
		return fmt.Errorf("arquivo de SHA-256 vazio: %s", url)
	}
	return sistema.VerificarSHA256(arquivo, campos[0])
}
