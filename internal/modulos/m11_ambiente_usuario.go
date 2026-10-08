package modulos

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

const dotfilesRepo = "https://github.com/vinicgobbi/Dotfiles.git"

// Prioriza o .desktop do pacote nativo; usa o do Flatpak como alternativa.
var solaarDesktop = []string{
	"/usr/share/applications/solaar.desktop",
	"/var/lib/flatpak/exports/share/applications/io.github.pwr_solaar.solaar.desktop",
}

var ambienteUsuario = &Modulo{
	ID:        "ambiente_usuario",
	Titulo:    "Configurar ambiente do usuário",
	Descricao: "Zsh, Oh My Zsh, Node (via mise) e dotfiles de shell (o tema visual fica com o Omarchy)",
	Executar:  configurarUsuario,
	Deps:      []string{"pacotes_base"},
}

// passoUsuario: passos de ambiente do usuário são best-effort: avisa e segue.
func passoUsuario(s *sistema.Sistema, script, descricao string, o sistema.Opts) error {
	ok, err := s.Tentar(script, o)
	if err != nil {
		return err
	}
	if !ok {
		s.UI.Aviso(fmt.Sprintf("Falha ao %s; seguindo com o resto.", descricao))
	}
	return nil
}

func contem(caminho, trecho string) bool {
	conteudo, err := os.ReadFile(caminho)
	return err == nil && strings.Contains(string(conteudo), trecho)
}

func configurarUsuario(s *sistema.Sistema) error {
	s.UI.Info(fmt.Sprintf("Aplicando configurações locais para %s...", s.Usuario()))
	if err := s.Run("usermod", "-aG", "docker", s.Usuario()); err != nil {
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
	passos := []struct {
		script, descricao string
		opts              sistema.Opts
	}{
		// Git Credential Manager
		{"git-credential-manager configure", "configurar o Git Credential Manager", sistema.Opts{}},
		{"git config --global credential.credentialStore secretservice", "definir o credentialStore do Git", sistema.Opts{}},
		// Clone+bootstrap dos dotfiles (Oh My Zsh, plugins, tema, fontes e
		// config do Solaar).
		{"rm -rf /tmp/dotfiles && git clone " + dotfilesRepo + " /tmp/dotfiles && bash /tmp/dotfiles/bootstrap.sh",
			"aplicar os Dotfiles", sistema.Opts{}},
		// Node via mise, que o Omarchy já usa por padrão (o próprio Omarchy
		// costuma já ter fixado uma versão no ~/.config/mise/config.toml; só
		// instalamos se não).
		{"mise which node &>/dev/null || mise use --global node@lts", "instalar o Node via mise", sistema.Opts{}},
	}
	for _, p := range passos {
		if err := passoUsuario(s, p.script, p.descricao, p.opts); err != nil {
			return err
		}
	}

	// Ativa o mise no zsh (o Omarchy só ativa no bash). Precisa rodar depois
	// do Oh My Zsh, que é quem cria/substitui o ~/.zshrc.
	zshrc := filepath.Join(s.Home(), ".zshrc")
	if !contem(zshrc, "mise activate zsh") {
		if err := s.Escrever(zshrc, "eval \"$(mise activate zsh)\"\n",
			sistema.Escrita{Anexar: true, DoUsuario: true}); err != nil {
			return err
		}
	}

	if err := configurarProjetos(s); err != nil {
		return err
	}
	if err := configurarAutostartSolaar(s); err != nil {
		return err
	}
	s.UI.Sucesso("Ambiente de usuário configurado.")
	return nil
}

// configurarProjetos cria o diretório de Projetos (XDG e bookmark do
// Nautilus), com o nome no idioma do usuário (o LANG vem do login shell dele).
func configurarProjetos(s *sistema.Sistema) error {
	res, err := s.ComoUsuario(`printf %s "$LANG"`, sistema.Opts{Capturar: true, Leitura: true, SemCheck: true})
	if err != nil {
		return err
	}
	nome := "Projects"
	if strings.HasPrefix(res.Saida, "pt_") {
		nome = "Projetos"
	}
	projetos := filepath.Join(s.Home(), nome)
	if err := s.CriarDirUsuario(projetos); err != nil {
		return err
	}
	if err := passoUsuario(s, sistema.Juntar("xdg-user-dirs-update", "--set", "PROJECTS", projetos),
		"registrar o diretório de projetos no XDG", sistema.Opts{}); err != nil {
		return err
	}

	bookmarks := filepath.Join(s.Home(), ".config/gtk-3.0/bookmarks")
	if !contem(bookmarks, "file://"+projetos) {
		return s.Escrever(bookmarks, "file://"+projetos+"\n", sistema.Escrita{Anexar: true, DoUsuario: true})
	}
	return nil
}

func configurarAutostartSolaar(s *sistema.Sistema) error {
	autostart := filepath.Join(s.Home(), ".config/autostart")
	if err := s.CriarDirUsuario(autostart); err != nil {
		return err
	}
	for _, origem := range solaarDesktop {
		conteudo, err := os.ReadFile(origem)
		if err != nil {
			continue
		}
		return s.Escrever(filepath.Join(autostart, filepath.Base(origem)), string(conteudo),
			sistema.Escrita{DoUsuario: true})
	}
	return nil
}
