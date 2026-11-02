package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var firefoxThunderbird = &Modulo{
	ID:        "firefox_thunderbird",
	Grupo:     GrupoAplicativos,
	Titulo:    "Firefox e Thunderbird",
	Descricao: "Firefox e Thunderbird pelos pacotes oficiais do Arch",
	Executar:  instalarFirefoxThunderbird,
}

func instalarFirefoxThunderbird(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Firefox e o Thunderbird...")
	if err := s.Pacman(s.Dados.Lista("firefox_thunderbird", "pacman")...); err != nil {
		return err
	}
	s.UI.Sucesso("Firefox e Thunderbird instalados.")
	return nil
}
