import re
from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_virt_manager(s):
    ui.info("Instalando virt-manager (QEMU/KVM + libvirt)...")

    if not re.search(r"vmx|svm", Path("/proc/cpuinfo").read_text()):
        ui.aviso("CPU sem suporte a virtualização por hardware (VT-x/AMD-V) ou não habilitado na "
                 "BIOS/UEFI — o KVM pode não funcionar.")

    s.pacman(*s.pacotes("virt_manager")["pacman"])

    s.run(["systemctl", "enable", "--now", "libvirtd"])

    for grupo in ("libvirt", "kvm"):
        if s.grupo_existe(grupo):
            s.run(["usermod", "-aG", grupo, s.usuario])

    ui.sucesso(f"virt-manager instalado. '{s.usuario}' foi adicionado aos grupos libvirt/kvm "
               "(é preciso reabrir a sessão para valer).")


MODULO = Modulo("virt_manager", "Instalar virt-manager (QEMU/KVM)",
                "QEMU/KVM, libvirt e a interface gráfica virt-manager para máquinas virtuais",
                instalar_virt_manager)
