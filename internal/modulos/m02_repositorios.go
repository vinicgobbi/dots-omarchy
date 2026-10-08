package modulos

import (
	"errors"
	"os"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

// Chave GPG oficial do Chaotic-AUR (aur.chaotic.cx) — repo binário que serve
// pacotes pré-compilados do AUR, evitando compilar tudo na mão.
const (
	chaoticAURKey = "3056513887B78AEB"
	pacmanConf    = "/etc/pacman.conf"
)

var repositorios = &Modulo{
	ID:        "repositorios",
	Titulo:    "Configurar repositórios",
	Descricao: "Habilita o Chaotic-AUR (pacotes da AUR pré-compilados) para o yay",
	Executar:  configurarRepositorios,
}

// configurarChaoticAUR segue o processo oficial (https://aur.chaotic.cx/):
// importa a chave, instala keyring+mirrorlist via pacman -U e inclui o
// mirrorlist no pacman.conf. Idempotente.
func configurarChaoticAUR(s *sistema.Sistema) error {
	conf, err := os.ReadFile(pacmanConf)
	if err != nil {
		return err
	}
	for _, linha := range strings.Split(string(conf), "\n") {
		if strings.HasPrefix(linha, "[chaotic-aur]") {
			s.UI.Info("Chaotic-AUR já configurado.")
			return nil
		}
	}

	s.UI.Info("Configurando o repositório binário Chaotic-AUR...")
	passos := [][]string{
		{"pacman-key", "--recv-key", chaoticAURKey, "--keyserver", "keyserver.ubuntu.com"},
		{"pacman-key", "--lsign-key", chaoticAURKey},
		{"pacman", "-U", "--noconfirm",
			"https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst",
			"https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst"},
	}
	for _, passo := range passos {
		if err := s.Run(passo...); err != nil {
			return err
		}
	}
	if err := s.Escrever(pacmanConf, "\n[chaotic-aur]\nInclude = /etc/pacman.d/chaotic-mirrorlist\n",
		sistema.Escrita{Anexar: true}); err != nil {
		return err
	}
	if err := s.Run("pacman", "-Sy"); err != nil {
		return err
	}
	s.UI.Sucesso("Chaotic-AUR configurado.")
	return nil
}

func configurarRepositorios(s *sistema.Sistema) error {
	s.UI.Info("Configurando repositórios (Chaotic-AUR + yay)...")

	// O mirrorlist do Omarchy já aponta para o CDN próprio dele
	// (stable-mirror.omarchy.org), que também serve o repositório [omarchy];
	// por isso não rodamos o reflector aqui. O yay já vem pré-instalado.
	if err := s.Run(append([]string{"pacman", "-Sy", "--needed", "--noconfirm"},
		s.Dados.Lista("repositorios", "pacman")...)...); err != nil {
		return err
	}
	if err := configurarChaoticAUR(s); err != nil {
		return err
	}
	if !sistema.ComandoExiste("yay") {
		return errors.New("yay não encontrado — deveria vir pré-instalado no Omarchy")
	}
	s.UI.Sucesso("Repositórios configurados.")
	return nil
}
