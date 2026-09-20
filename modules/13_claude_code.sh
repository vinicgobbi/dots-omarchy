#!/usr/bin/env bash

instalar_claude_code() {
    # O Omarchy já instala e mantém o Claude Code via mise (omarchy-mise-install
    # claude), disponível no PATH do usuário via shim. Rodar o instalador oficial
    # por cima disso criaria uma segunda cópia concorrendo no PATH; então, se já
    # existir para o usuário-alvo (por qualquer via), pulamos.
    if executar_como_usuario "command -v claude" &>/dev/null; then
        aviso "Claude Code já está instalado para $USER_NAME (via mise, no Omarchy); pulando reinstalação."
        return
    fi

    info "Instalando o Claude Code..."

    executar_como_usuario "
  curl -fsSL https://claude.ai/install.sh | bash
"
    sucesso "Claude Code instalado para $USER_NAME."
}

registrar_modulo "claude_code" "Instalar Claude Code" \
    "CLI oficial da Anthropic para desenvolvimento assistido por IA no terminal" \
    "instalar_claude_code"
