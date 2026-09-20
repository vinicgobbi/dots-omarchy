#!/usr/bin/env bash

instalar_tailscale() {
    info "Instalando o Tailscale..."

    # tailscale é pacote oficial do repo "extra" — mesmo pacote que
    # "omarchy install service tailscale" usa.
    pacman -S --needed --noconfirm tailscale

    systemctl enable --now tailscaled
    tailscale set --operator="$USER_NAME"

    integrar_tailscale_omarchy

    sucesso "Tailscale instalado e '$USER_NAME' definido como operator."
}

# Replica o que "omarchy-install-service-tailscale" faz: widget na barra,
# atalho de webapp pro admin console e recebimento de Taildrop em
# ~/Downloads. Usa USER_NAME (em vez de $USER) e roda cada passo como o
# usuário-alvo, já que este script inteiro roda como root.
# Cada etapa é best-effort (não aborta o módulo): a sessão gráfica do
# usuário-alvo pode não estar ativa ainda nesse momento da instalação.
integrar_tailscale_omarchy() {
    local runtime_dir="/run/user/$(id -u "$USER_NAME")"

    if executar_como_usuario "XDG_RUNTIME_DIR='$runtime_dir' systemctl --user enable --now omarchy-tailscale-receive.service" &>/dev/null; then
        info "Recebimento de Taildrop habilitado em ~/Downloads."
    else
        aviso "Não foi possível habilitar o recebimento de Taildrop agora (sessão gráfica de '$USER_NAME' inativa?); rode depois de logar: systemctl --user enable --now omarchy-tailscale-receive.service"
    fi

    if executar_como_usuario "omarchy-plugin-enable omarchy.tailscale" &>/dev/null; then
        info "Tailscale adicionado à barra do Omarchy."
    else
        aviso "Não foi possível adicionar o Tailscale à barra do Omarchy agora."
    fi

    criar_webapp_tailscale_admin
}

# Extraída à parte porque o módulo "limpeza" também chama isto: ele apaga
# todos os webapps do Omarchy (omarchy-webapp-remove-all) e, se este módulo
# tiver rodado na mesma execução, precisa recriar o atalho perdido.
criar_webapp_tailscale_admin() {
    executar_como_usuario "omarchy-webapp-install 'Tailscale' 'https://login.tailscale.com/admin/machines' https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/png/tailscale-light.png" &>/dev/null \
        || aviso "Não foi possível criar o atalho de webapp do Tailscale Admin Console."
}

registrar_modulo "tailscale" "Instalar Tailscale" \
    "Tailscale (VPN mesh) com widget na barra e webapp do admin console" \
    "instalar_tailscale"
