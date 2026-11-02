package modulos

import (
	"fmt"
	"os/exec"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var shell = &Modulo{
	ID:        "shell",
	Grupo:     GrupoTerminal,
	Titulo:    "Zsh e dotfiles",
	Descricao: "Zsh como shell padrão e os dotfiles pessoais: Oh My Zsh, plugins e fontes (o tema visual fica com o Omarchy)",
	Executar:  configurarShell,
}

func configurarShell(s *sistema.Sistema) error {
	s.UI.Info(fmt.Sprintf("Configurando o Zsh e os dotfiles para %s...", s.Usuario()))
	if err := s.Pacman(s.Dados.Lista("shell", "pacman")...); err != nil {
		return err
	}

	// chsh só aceita contas presentes de fato em /etc/passwd; contas de
	// domínio (AD/LDAP via sssd/winbind) resolvem por NSS mas não estão no
	// arquivo, e o chsh recusa. A resolução restrita à fonte "files" (local)
	// distingue os dois casos.
	if s.Ok("getent", "-s", "files", "passwd", s.Usuario()) {
		zsh, err := exec.LookPath("zsh")
		if err != nil {
			zsh = "/usr/bin/zsh"
		}
		if err := s.Run("chsh", "-s", zsh, s.Usuario()); err != nil {
			return err
		}
	} else {
		s.UI.Aviso(fmt.Sprintf("'%s' é uma conta de domínio (AD/LDAP), não local: pulando 'chsh' "+
			"(não dá para trocar o shell padrão por aqui nesse caso).", s.Usuario()))
	}

	// O tema (GTK, ícones, Nautilus etc.) já é gerido pelo próprio Omarchy
	// ("omarchy theme"), então não aplicamos gsettings/adw-gtk3/Yaru aqui.
	if err := s.Usr(inicioScriptTemp + clonarFixado(dotfilesRepo, dotfilesCommit, `"$dir"`) +
		`bash "$dir/bootstrap.sh"` + "\n"); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("Zsh e dotfiles configurados para %s.", s.Usuario()))
	return nil
}
