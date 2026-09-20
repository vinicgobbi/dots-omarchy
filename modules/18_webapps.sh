#!/usr/bin/env bash

# Webapps criados (nome|url|ícone). Os ícones vêm do dashboard-icons (mesmo
# CDN usado no atalho do Tailscale).
ICONES_WEBAPP="https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png"
WEBAPPS=(
    "YouTube|https://www.youtube.com|$ICONES_WEBAPP/youtube.png"
    "WhatsApp|https://web.whatsapp.com|$ICONES_WEBAPP/whatsapp.png"
    "Gmail|https://mail.google.com|$ICONES_WEBAPP/gmail.png"
    "Netflix|https://www.netflix.com|$ICONES_WEBAPP/netflix.png"
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

    # omarchy-webapp-remove-all não distingue os webapps de fábrica do atalho
    # que o módulo "tailscale" criou — se ele foi selecionado nesta execução,
    # recria o atalho perdido.
    local _i
    for ((_i = 0; _i < ${#MOD_IDS[@]}; _i++)); do
        if [[ "${MOD_IDS[_i]}" == "tailscale" && "${MOD_SEL[_i]}" == "1" ]]; then
            criar_webapp_tailscale_admin
            info "Atalho de webapp do Tailscale Admin Console recriado."
            break
        fi
    done

    sucesso "Webapps configurados."
}

registrar_modulo "webapps" "Criar webapps" \
    "Apaga os webapps de fábrica do Omarchy e cria YouTube, WhatsApp, Gmail e Netflix" \
    "criar_webapps"
