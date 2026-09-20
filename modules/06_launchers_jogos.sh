#!/usr/bin/env bash

instalar_launchers_jogos() {
    instalar_jogos_nativos

    info "Instalando launchers de jogos via Flatpak..."

    flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
    if [[ "${#FLATPAKS_JOGOS[@]}" -gt 0 ]]; then
        flatpak install -y flathub "${FLATPAKS_JOGOS[@]}"
    fi
    sucesso "Launchers de jogos instalados."
}

# Steam (multilib) e Heroic (repositório [omarchy]) têm pacote nativo pronto
# — mesmo caminho que "omarchy install gaming steam"/"heroic" usam —
# incluindo a detecção e instalação dos drivers gráficos lib32 (Vulkan
# Intel/AMD, NVIDIA) certos pra essa máquina.
instalar_jogos_nativos() {
    info "Instalando Steam e Heroic (pacotes nativos, como o instalador do Omarchy faria)..."
    pacman -S --needed --noconfirm steam heroic-games-launcher-bin
    comando_existe omarchy-install-gaming-gpu-lib32 && omarchy-install-gaming-gpu-lib32
    sucesso "Steam e Heroic instalados nativamente."
}

registrar_modulo "launchers_jogos" "Instalar launchers de jogos" \
    "Steam e Heroic nativos; ProtonPlus e PrismLauncher via Flatpak" \
    "instalar_launchers_jogos" "pacotes_base"
