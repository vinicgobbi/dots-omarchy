package modulos

import (
	"path/filepath"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var pacotesBase = &Modulo{
	ID:        "pacotes_base",
	Titulo:    "Instalar pacotes base",
	Descricao: "Docker, VSCode, PHP, Zsh e ferramentas de SQL Server",
	Executar:  instalarPacotesBase,
	Deps:      []string{"repositorios"},
}

func instalarPacotesBase(s *sistema.Sistema) error {
	s.UI.Info("Instalando pacotes base e utilitários...")
	// Docker, PHP/composer e unixODBC são oficiais (repo extra/core). O
	// repositório [omarchy] já traz o visual-studio-code-bin pré-compilado
	// (mesmo pacote que "omarchy install editor vscode" usa); msodbcsql e
	// mssql-tools só existem na AUR (resolvidos pelo Chaotic-AUR, sem compilar).
	if err := s.Pacman(s.Dados.Lista("pacotes_base", "pacman")...); err != nil {
		return err
	}
	if err := s.Aur(s.Dados.Lista("pacotes_base", "aur")...); err != nil {
		return err
	}
	if err := aplicarConfigOmarchyVSCode(s); err != nil {
		return err
	}

	// O pacote da AUR instala em /opt/mssql-tools (sem versão).
	if err := s.Escrever("/etc/profile.d/mssql-tools.sh", "export PATH=\"$PATH:/opt/mssql-tools/bin\"\n",
		sistema.Escrita{}); err != nil {
		return err
	}
	s.UI.Sucesso("Pacotes e MS SQL instalados.")
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
