#!/usr/bin/env bash

configurar_solaar() {
    info "Instalando Solaar..."

    if pacman -Si solaar &>/dev/null; then
        pacman -S --needed --noconfirm solaar
        sucesso "Solaar instalado via pacote nativo (pacman)."
    else
        aviso "Pacote nativo do Solaar não encontrado; instalando via Flatpak."
        flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo
        flatpak install -y flathub io.github.pwr_solaar.solaar
        sucesso "Solaar instalado via Flatpak."
    fi

    getent group plugdev >/dev/null || groupadd plugdev
    usermod -aG plugdev "$USER_NAME"

    info "Configurando regras UDEV para o Solaar..."
    curl -sL https://raw.githubusercontent.com/pwr-Solaar/Solaar/master/rules.d-uinput/42-logitech-unify-permissions.rules -o /etc/udev/rules.d/42-logitech-unify-permissions.rules

    udevadm control --reload-rules
    udevadm trigger --subsystem-match=usb
    udevadm trigger --subsystem-match=hidraw

    sucesso "Regras UDEV do Solaar configuradas."
}

registrar_modulo "solaar" "Instalar e configurar Solaar" \
    "Instala o Solaar (nativo, com fallback para Flatpak) e as regras UDEV do receptor Logitech Unifying" \
    "configurar_solaar" "pacotes_base"
