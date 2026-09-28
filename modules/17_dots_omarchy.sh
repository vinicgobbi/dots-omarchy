#!/usr/bin/env bash

# Copia um arquivo de dots/ para ~/<destino> como o usuário-alvo,
# guardando o original em <arquivo>.bak-post-omarchy na primeira vez que ele
# difere do que vamos instalar (para nunca perder customização anterior).
# Uso: instalar_dot <origem> <destino_relativo_ao_home> [modo]
instalar_dot() {
    local origem="$1" destino="$2" modo="${3:-0644}"
    local alvo
    alvo="$(getent passwd "$USER_NAME" | cut -d: -f6)/$destino"

    if [[ -f "$alvo" ]]; then
        cmp -s "$origem" "$alvo" && return
        [[ -e "$alvo.bak-post-omarchy" ]] || cp -p "$alvo" "$alvo.bak-post-omarchy"
    fi

    executar_como_usuario "mkdir -p '$(dirname "$alvo")'"
    install -o "$USER_NAME" -g "$(id -gn "$USER_NAME")" -m "$modo" "$origem" "$alvo"
    info "Instalado ~/$destino"
}

aplicar_dots_omarchy() {
    info "Aplicando dots do Omarchy para $USER_NAME..."

    local arquivo rel
    for arquivo in "$SCRIPT_DIR"/dots/hypr/*.lua; do
        instalar_dot "$arquivo" ".config/hypr/$(basename "$arquivo")"
    done

    instalar_dot "$SCRIPT_DIR/dots/omarchy/shell.json" ".config/omarchy/shell.json"
    # Screensaver personalizado (o do Omarchy fica como .bak-post-omarchy).
    instalar_dot "$SCRIPT_DIR/dots/omarchy/branding/screensaver.txt" ".config/omarchy/branding/screensaver.txt"
    # Extensões (dots/omarchy/extensions/), se houver alguma.
    while IFS= read -r -d '' arquivo; do
        rel="${arquivo#"$SCRIPT_DIR/dots/omarchy/"}"
        [[ "$(basename "$arquivo")" == ".gitkeep" ]] && continue
        instalar_dot "$arquivo" ".config/omarchy/$rel"
    done < <(find "$SCRIPT_DIR/dots/omarchy/extensions" -type f -print0)

    # Scripts auxiliares em ~/.local/bin: fullscreen-dnd (chamado pelo
    # hypr/autostart.lua) e solaar-volume (chamado pelo solaar/rules.yaml).
    for arquivo in "$SCRIPT_DIR"/dots/bin/*; do
        instalar_dot "$arquivo" ".local/bin/$(basename "$arquivo")" 0755
    done

    # Solaar (MX Master 3S): config.yaml desvia o botão de gesto e a thumb wheel,
    # rules.yaml mapeia gestos/roda para comandos via Execute. O Execute não
    # expande ~, então o rules.yaml traz o caminho absoluto de /home/vinicius,
    # trocado aqui pelo home do usuário-alvo.
    instalar_dot "$SCRIPT_DIR/dots/solaar/config.yaml" ".config/solaar/config.yaml"
    local home_alvo regras
    home_alvo="$(getent passwd "$USER_NAME" | cut -d: -f6)"
    regras="$(mktemp)"
    sed "s|/home/vinicius/|$home_alvo/|g" "$SCRIPT_DIR/dots/solaar/rules.yaml" > "$regras"
    instalar_dot "$regras" ".config/solaar/rules.yaml"
    rm -f "$regras"

    # Memória global do Claude Code: não é versionada (ver dots/claude/README.md),
    # então só é instalada se alguém tiver colocado o arquivo lá.
    if [[ -f "$SCRIPT_DIR/dots/claude/CLAUDE.md" ]]; then
        instalar_dot "$SCRIPT_DIR/dots/claude/CLAUDE.md" ".claude/CLAUDE.md"
    else
        aviso "dots/claude/CLAUDE.md não encontrado; pulando a memória global do Claude Code."
    fi

    sucesso "Dots do Omarchy aplicados. Recarregue o Hyprland (super+shift+r, ou faça logout/login) para ver as mudanças."
}

registrar_modulo "dots_omarchy" "Aplicar dots do Omarchy" \
    "Copia a config do Hyprland, do omarchy-shell, o screensaver, o Solaar, os scripts de ~/.local/bin e o CLAUDE.md global, se houver (com backup do que existia)" \
    "aplicar_dots_omarchy"
