#!/usr/bin/env bash

# Copia um arquivo de dots/ para ~/.config/<destino> como o usuário-alvo,
# guardando o original em <arquivo>.bak-post-omarchy na primeira vez que ele
# difere do que vamos instalar (para nunca perder customização anterior).
# Uso: instalar_dot <origem> <destino_relativo_a_.config>
instalar_dot() {
    local origem="$1" destino="$2"
    local alvo
    alvo="$(getent passwd "$USER_NAME" | cut -d: -f6)/.config/$destino"

    if [[ -f "$alvo" ]]; then
        cmp -s "$origem" "$alvo" && return
        [[ -e "$alvo.bak-post-omarchy" ]] || cp -p "$alvo" "$alvo.bak-post-omarchy"
    fi

    executar_como_usuario "mkdir -p '$(dirname "$alvo")'"
    install -o "$USER_NAME" -g "$(id -gn "$USER_NAME")" -m 0644 "$origem" "$alvo"
    info "Instalado ~/.config/$destino"
}

aplicar_dots_omarchy() {
    info "Aplicando dots do Omarchy para $USER_NAME..."

    local arquivo rel
    for arquivo in "$SCRIPT_DIR"/dots/hypr/*.lua; do
        instalar_dot "$arquivo" "hypr/$(basename "$arquivo")"
    done

    instalar_dot "$SCRIPT_DIR/dots/omarchy/shell.json" "omarchy/shell.json"
    # Extensões (dots/omarchy/extensions/), se houver alguma.
    while IFS= read -r -d '' arquivo; do
        rel="${arquivo#"$SCRIPT_DIR/dots/omarchy/"}"
        [[ "$(basename "$arquivo")" == ".gitkeep" ]] && continue
        instalar_dot "$arquivo" "omarchy/$rel"
    done < <(find "$SCRIPT_DIR/dots/omarchy/extensions" -type f -print0)

    instalar_plugins_omarchy

    sucesso "Dots do Omarchy aplicados. Recarregue o Hyprland (super+shift+r, ou faça logout/login) para ver as mudanças."
}

# Cada etapa é best-effort: "omarchy plugin clone" precisa de rede e o
# omarchy-shell pode não estar rodando (sessão gráfica ainda inativa).
instalar_plugins_omarchy() {
    local url
    for url in "${OMARCHY_PLUGINS[@]}"; do
        info "Plugin do Omarchy: $url"
        executar_como_usuario "omarchy plugin clone '$url'" \
            || aviso "Não foi possível instalar o plugin $url; rode depois: omarchy plugin clone $url"
    done
}

registrar_modulo "dots_omarchy" "Aplicar dots do Omarchy" \
    "Copia a config do Hyprland e do omarchy-shell (com backup do que existia) e instala os plugins da barra" \
    "aplicar_dots_omarchy"
