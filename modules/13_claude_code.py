from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_claude_code(s):
    # O Omarchy já instala e mantém o Claude Code via mise (omarchy-mise-install
    # claude), disponível no PATH do usuário via shim. Rodar o instalador oficial
    # por cima disso criaria uma segunda cópia concorrendo no PATH; então, se já
    # existir para o usuário-alvo (por qualquer via), pulamos.
    if s.ok("command -v claude", como_usuario=True):
        ui.aviso(f"Claude Code já está instalado para {s.usuario} (via mise, no Omarchy); pulando reinstalação.")
        return

    ui.info("Instalando o Claude Code...")
    s.como_usuario("curl -fsSL https://claude.ai/install.sh | bash")
    ui.sucesso(f"Claude Code instalado para {s.usuario}.")


MODULO = Modulo("claude_code", "Instalar Claude Code",
                "CLI oficial da Anthropic para desenvolvimento assistido por IA no terminal",
                instalar_claude_code)
