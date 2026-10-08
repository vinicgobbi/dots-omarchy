package modulos

import (
	"fmt"
	"os"
	"regexp"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var virtManager = &Modulo{
	ID:        "virt_manager",
	Titulo:    "Instalar virt-manager (QEMU/KVM)",
	Descricao: "QEMU/KVM, libvirt e a interface gráfica virt-manager para máquinas virtuais",
	Executar:  instalarVirtManager,
}

func instalarVirtManager(s *sistema.Sistema) error {
	s.UI.Info("Instalando virt-manager (QEMU/KVM + libvirt)...")

	cpuinfo, _ := os.ReadFile("/proc/cpuinfo")
	if !regexp.MustCompile(`vmx|svm`).Match(cpuinfo) {
		s.UI.Aviso("CPU sem suporte a virtualização por hardware (VT-x/AMD-V) ou não habilitado na " +
			"BIOS/UEFI — o KVM pode não funcionar.")
	}

	if err := s.Pacman(s.Dados.Lista("virt_manager", "pacman")...); err != nil {
		return err
	}
	if err := s.Run("systemctl", "enable", "--now", "libvirtd"); err != nil {
		return err
	}
	for _, grupo := range []string{"libvirt", "kvm"} {
		if sistema.GrupoExiste(grupo) {
			if err := s.Run("usermod", "-aG", grupo, s.Usuario()); err != nil {
				return err
			}
		}
	}
	s.UI.Sucesso(fmt.Sprintf("virt-manager instalado. '%s' foi adicionado aos grupos libvirt/kvm "+
		"(é preciso reabrir a sessão para valer).", s.Usuario()))
	return nil
}
