package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var claudeCode = &Modulo{
	ID:        "claude_code",
	Titulo:    "Instalar Claude Code",
	Descricao: "CLI oficial da Anthropic para desenvolvimento assistido por IA no terminal",
	Executar:  instalarClaudeCode,
}

func instalarClaudeCode(s *sistema.Sistema) error {
	// O Omarchy já instala e mantém o Claude Code via mise (omarchy-mise-install
	// claude), disponível no PATH do usuário via shim. Rodar o instalador oficial
	// por cima disso criaria uma segunda cópia concorrendo no PATH; então, se já
	// existir para o usuário-alvo (por qualquer via), pulamos.
	if s.OkUsuario("command -v claude") {
		s.UI.Aviso(fmt.Sprintf("Claude Code já está instalado para %s (via mise, no Omarchy); pulando reinstalação.", s.Usuario()))
		return nil
	}

	s.UI.Info("Instalando o Claude Code...")
	if err := s.Usr("curl -fsSL https://claude.ai/install.sh | bash"); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("Claude Code instalado para %s.", s.Usuario()))
	return nil
}
