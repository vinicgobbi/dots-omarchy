package modulos

import (
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

const (
	gcmRepo = "git-ecosystem/git-credential-manager"
	gcmDir  = "/usr/local/gcm"
)

var gitGCM = &Modulo{
	ID:     "git_gcm",
	Grupo:  GrupoDesenvolvimento,
	Titulo: "Git Credential Manager",
	Descricao: "Instala o Git Credential Manager (última release, com SHA-256 conferido) como credential " +
		"helper do Git, guardando as credenciais no chaveiro do sistema",
	Executar: instalarGCM,
}

// assetGCM escolhe o .tar.gz linux-x64 (sem os símbolos).
func assetGCM(nome string) bool {
	return strings.HasSuffix(nome, ".tar.gz") && strings.Contains(nome, "linux-x64") &&
		!strings.Contains(nome, "symbols")
}

func instalarGCM(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Git Credential Manager...")
	if err := s.Pacman(s.Dados.Lista("git_gcm", "pacman")...); err != nil {
		return err
	}
	asset, err := ultimoAsset(s, gcmRepo, assetGCM)
	if err != nil {
		return err
	}
	pacote := s.Temp("gcm.tar.gz")
	if err := baixarAsset(s, asset, pacote); err != nil {
		return err
	}
	if err := s.Run("mkdir", "-p", gcmDir); err != nil {
		return err
	}
	if err := s.Run("tar", "-xzf", pacote, "-C", gcmDir); err != nil {
		return err
	}
	if err := s.Link(filepath.Join(gcmDir, "git-credential-manager"), "/usr/local/bin/git-credential-manager"); err != nil {
		return err
	}
	if err := s.Usr("git-credential-manager configure && git config --global credential.credentialStore secretservice"); err != nil {
		return err
	}
	s.UI.Sucesso("Git Credential Manager instalado e configurado para " + s.Usuario() + ".")
	return nil
}
