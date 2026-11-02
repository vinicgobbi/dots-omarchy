package modulos

import (
	"path/filepath"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var vscode = &Modulo{
	ID:        "vscode",
	Grupo:     GrupoDesenvolvimento,
	Titulo:    "Visual Studio Code",
	Descricao: "VSCode pelo pacote do repositório [omarchy], com o tema atual do Omarchy",
	Executar:  instalarVSCode,
}

func instalarVSCode(s *sistema.Sistema) error {
	s.UI.Info("Instalando o VSCode...")
	// O repositório [omarchy] já traz o visual-studio-code-bin pré-compilado
	// (mesmo pacote que "omarchy install editor vscode" usa).
	if err := s.Pacman(s.Dados.Lista("vscode", "pacman")...); err != nil {
		return err
	}
	if err := aplicarConfigOmarchyVSCode(s); err != nil {
		return err
	}
	s.UI.Sucesso("VSCode instalado.")
	return nil
}

// aplicarConfigOmarchyVSCode replica o que "omarchy-install-editor-vscode" faz
// para o usuário-alvo (senha via gnome-libsecret, autoupdate desligado — o
// Omarchy já atualiza o VSCode via pacman — e o tema atual do Omarchy aplicado
// no editor), sem o lançamento automático da GUI ao final, que assume uma
// sessão gráfica interativa e não faz sentido no meio de uma instalação em lote.
func aplicarConfigOmarchyVSCode(s *sistema.Sistema) error {
	if err := s.Escrever(filepath.Join(s.Home(), ".vscode/argv.json"),
		"{\n  \"password-store\":\"gnome-libsecret\"\n}\n", sistema.Escrita{DoUsuario: true}); err != nil {
		return err
	}
	if err := s.Escrever(filepath.Join(s.Home(), ".config/Code/User/settings.json"),
		"{\n  \"update.mode\": \"none\"\n}\n", sistema.Escrita{DoUsuario: true}); err != nil {
		return err
	}
	return s.Usr("omarchy-theme-set-vscode")
}
