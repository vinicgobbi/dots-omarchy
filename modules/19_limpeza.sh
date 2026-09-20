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

    sucesso "Instalação finalizada com sucesso!"
    info "Recomenda-se reiniciar a máquina para aplicar as mudanças de grupo e kernel."
}

registrar_modulo "limpeza" "Limpeza final" \
    "Remove pacotes órfãos, limpa o cache do pacman e remove o sudo temporário do yay" \
    "limpeza_final"
