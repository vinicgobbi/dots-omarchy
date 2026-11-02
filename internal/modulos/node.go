package modulos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var node = &Modulo{
	ID:        "node",
	Grupo:     GrupoDesenvolvimento,
	Titulo:    "Node.js (mise)",
	Descricao: "Node LTS pelo mise, que o Omarchy já usa por padrão, com o mise ativo também no zsh",
	Executar:  instalarNode,
}

func contem(caminho, trecho string) bool {
	conteudo, err := os.ReadFile(caminho)
	return err == nil && strings.Contains(string(conteudo), trecho)
}

func instalarNode(s *sistema.Sistema) error {
	s.UI.Info(fmt.Sprintf("Instalando o Node para %s...", s.Usuario()))
	// O próprio Omarchy costuma já ter fixado uma versão no
	// ~/.config/mise/config.toml; só instalamos se não.
	if err := s.Usr("mise which node &>/dev/null || mise use --global node@lts"); err != nil {
		return err
	}

	// Ativa o mise no zsh (o Omarchy só ativa no bash). Roda depois do módulo
	// shell, porque o Oh My Zsh é quem cria/substitui o ~/.zshrc; sem ele, não
	// há .zshrc para mexer.
	zshrc := filepath.Join(s.Home(), ".zshrc")
	if _, err := os.Stat(zshrc); err == nil && !contem(zshrc, "mise activate zsh") {
		if err := s.Escrever(zshrc, "eval \"$(mise activate zsh)\"\n",
			sistema.Escrita{Anexar: true, DoUsuario: true}); err != nil {
			return err
		}
	}
	s.UI.Sucesso(fmt.Sprintf("Node instalado para %s.", s.Usuario()))
	return nil
}
