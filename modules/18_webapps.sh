#!/usr/bin/env bash

# Webapps criados (nome|url|ícone). Os ícones vêm do dashboard-icons (mesmo
# CDN usado no atalho do Tailscale), exceto os que ficam em assets/icons.
ICONES_WEBAPP="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png"
WEBAPPS=(
    "YouTube|https://www.youtube.com|$ICONES_WEBAPP/youtube.png"
    "WhatsApp|https://web.whatsapp.com|$ICONES_WEBAPP/whatsapp.png"
    "Gmail|https://mail.google.com|$ICONES_WEBAPP/gmail.png"
    "Netflix|https://www.netflix.com|$ICONES_WEBAPP/netflix.png"
    "Tailscale|https://login.tailscale.com/admin/machines|$ICONES_WEBAPP/tailscale-light.png"
    "Twitch|https://www.twitch.tv|$ICONES_WEBAPP/twitch.png"
    "Github|https://github.com|$ICONES_WEBAPP/github-light.png"
    "GLPI|https://sac.faesa.br|$SCRIPT_DIR/assets/icons/glpi.svg"
)

criar_webapps() {
    info "Removendo os webapps de fábrica do Omarchy (omarchy-webapp-remove-all)..."
    executar_como_usuario "omarchy-webapp-remove-all" \
        || aviso "Falha ao remover os webapps do Omarchy."

    local item nome url icone
    for item in "${WEBAPPS[@]}"; do
        IFS='|' read -r nome url icone <<<"$item"
        if executar_como_usuario "omarchy-webapp-install '$nome' '$url' '$icone'" &>/dev/null; then
            info "Webapp criado: $nome"
        else
            aviso "Não foi possível criar o webapp $nome."
        fi
    done

    sucesso "Webapps configurados."
}

registrar_modulo "webapps" "Criar webapps" \
    "Apaga os webapps de fábrica do Omarchy e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, GitHub, GLPI e o admin console do Tailscale" \
    "criar_webapps"
