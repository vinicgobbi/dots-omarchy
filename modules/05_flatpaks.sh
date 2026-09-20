#!/usr/bin/env bash

instalar_flatpaks() {
    # O Spotify tem pacote nativo pré-compilado no repositório [omarchy]
    # (mesmo que "omarchy install service spotify" usa), então instalamos via
    # pacman em vez de duplicar via Flatpak.
    info "Instalando Spotify (pacote nativo do repositório do Omarchy)..."
    pacman -S --needed --noconfirm spotify

    info "Configurando Flatpak e instalando apps..."

    flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
    if [[ "${#FLATPAKS[@]}" -gt 0 ]]; then
        flatpak install -y flathub "${FLATPAKS[@]}"
    fi
    sucesso "Aplicativos Flatpak instalados."
}

registrar_modulo "flatpaks" "Instalar Flatpaks" \
    "Spotify nativo e a lista de aplicativos Flatpak (Postman, Obsidian, etc.)" \
    "instalar_flatpaks" "pacotes_base"
