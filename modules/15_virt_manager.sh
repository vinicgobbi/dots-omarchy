#!/usr/bin/env bash

instalar_virt_manager() {
    info "Instalando virt-manager (QEMU/KVM + libvirt)..."

    if ! grep -qE 'vmx|svm' /proc/cpuinfo; then
        aviso "CPU sem suporte a virtualização por hardware (VT-x/AMD-V) ou não habilitado na BIOS/UEFI — o KVM pode não funcionar."
    fi

    pacman -S --needed --noconfirm qemu-desktop libvirt virt-manager

    systemctl enable --now libvirtd

    local grupo
    for grupo in libvirt kvm; do
        getent group "$grupo" >/dev/null && usermod -aG "$grupo" "$USER_NAME"
    done

    sucesso "virt-manager instalado. '$USER_NAME' foi adicionado aos grupos libvirt/kvm (é preciso reabrir a sessão para valer)."
}

registrar_modulo "virt_manager" "Instalar virt-manager (QEMU/KVM)" \
    "QEMU/KVM, libvirt e a interface gráfica virt-manager para máquinas virtuais" \
    "instalar_virt_manager"
