#!/usr/bin/env bash

limpeza_final() {
    info "Limpando o sistema..."
    local orfaos
    orfaos=$(pacman -Qtdq 2>/dev/null || true)
    [[ -n "$orfaos" ]] && pacman -Rns --noconfirm $orfaos
    pacman -Sc --noconfirm

    if [[ -f /etc/sudoers.d/99-post-omarchy-aur ]]; then
        rm -f /etc/sudoers.d/99-post-omarchy-aur
        info "Regra temporária de sudo sem senha para o pacman (liberada para o yay) removida."
    fi
    rm -rf /tmp/* 2>/dev/null || true

    info "Removendo webapps do Omarchy (omarchy-webapp-remove-all)..."
    executar_como_usuario "omarchy-webapp-remove-all" \
        || aviso "Falha ao remover os webapps do Omarchy."

    # omarchy-webapp-remove-all não distingue os webapps de fábrica do
    # atalho que o próprio módulo "tailscale" acabou de criar (ver
    # criar_webapp_tailscale_admin, em modules/07_tailscale.sh) — se ele foi
    # selecionado nesta execução, recria o atalho perdido.
    local _i
    for ((_i = 0; _i < ${#MOD_IDS[@]}; _i++)); do
        if [[ "${MOD_IDS[_i]}" == "tailscale" && "${MOD_SEL[_i]}" == "1" ]]; then
            criar_webapp_tailscale_admin
            info "Atalho de webapp do Tailscale Admin Console recriado."
            break
        fi
    done

    sucesso "Instalação finalizada com sucesso!"
    info "Recomenda-se reiniciar a máquina para aplicar as mudanças de grupo e kernel."
}

registrar_modulo "limpeza" "Limpeza final" \
    "Remove pacotes órfãos, limpa o cache do pacman e apaga todos os atalhos de webapp do Omarchy" \
    "limpeza_final"
