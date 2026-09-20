#!/usr/bin/env bash
# ==========================================
# Detecção de sistema operacional
# ==========================================
# Este projeto é exclusivo do Omarchy (Arch por baixo: pacman, yay
# pré-instalado, Hyprland). Define, ao final de detectar_sistema:
#   ID / PRETTY_NAME vindos de /etc/os-release

detectar_sistema() {
    . /etc/os-release

    if [[ "$ID" != "omarchy" ]]; then
        erro "Sistema não suportado: $PRETTY_NAME. Este script é exclusivo do Omarchy (ID=omarchy). Para outras distros, use o projeto post_install."
    fi
}

mensagem_sistema_detectado() {
    info "Sistema detectado: $PRETTY_NAME (Omarchy — Arch por baixo, pacman + Chaotic-AUR/yay)"
}
