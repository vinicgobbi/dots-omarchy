package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var flatpaks = &Modulo{
	ID:        "flatpaks",
	Titulo:    "Instalar Flatpaks",
	Descricao: "Spotify nativo e a lista de aplicativos Flatpak (Postman, Obsidian, etc.)",
	Executar:  instalarFlatpaks,
	Deps:      []string{"pacotes_base"},
}

func instalarFlatpaks(s *sistema.Sistema) error {
	// O Spotify tem pacote nativo pré-compilado no repositório [omarchy]
	// (mesmo que "omarchy install service spotify" usa), então instalamos via
	// pacman em vez de duplicar via Flatpak.
	s.UI.Info("Instalando Spotify (pacote nativo do repositório do Omarchy)...")
	if err := s.Pacman(s.Dados.Lista("flatpaks", "pacman")...); err != nil {
		return err
	}

	s.UI.Info("Configurando Flatpak e instalando apps...")
	if err := s.Flatpak(s.Dados.Flatpaks.Apps...); err != nil {
		return err
	}
	s.UI.Sucesso("Aplicativos Flatpak instalados.")
	return nil
}
