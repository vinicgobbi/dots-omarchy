package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var launchersJogos = &Modulo{
	ID:        "launchers_jogos",
	Titulo:    "Instalar launchers de jogos",
	Descricao: "Steam e Heroic nativos; ProtonPlus e PrismLauncher via Flatpak",
	Executar:  instalarLaunchersJogos,
	Deps:      []string{"pacotes_base"},
}

func instalarLaunchersJogos(s *sistema.Sistema) error {
	if err := instalarJogosNativos(s); err != nil {
		return err
	}
	s.UI.Info("Instalando launchers de jogos via Flatpak...")
	if err := s.Flatpak(s.Dados.Flatpaks.Jogos...); err != nil {
		return err
	}
	s.UI.Sucesso("Launchers de jogos instalados.")
	return nil
}

// instalarJogosNativos: Steam (multilib) e Heroic (repositório [omarchy]) têm
// pacote nativo pronto — mesmo caminho que "omarchy install gaming
// steam"/"heroic" usam — incluindo a detecção e instalação dos drivers gráficos
// lib32 (Vulkan Intel/AMD, NVIDIA) certos pra essa máquina.
func instalarJogosNativos(s *sistema.Sistema) error {
	s.UI.Info("Instalando Steam e Heroic (pacotes nativos, como o instalador do Omarchy faria)...")
	if err := s.Pacman(s.Dados.Lista("launchers_jogos", "pacman")...); err != nil {
		return err
	}
	if sistema.ComandoExiste("omarchy-install-gaming-gpu-lib32") {
		if err := s.Run("omarchy-install-gaming-gpu-lib32"); err != nil {
			return err
		}
	}
	s.UI.Sucesso("Steam e Heroic instalados nativamente.")
	return nil
}
