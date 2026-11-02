package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var golang = &Modulo{
	ID:        "go",
	Grupo:     GrupoDesenvolvimento,
	Titulo:    "Go",
	Descricao: "Go pelo pacote oficial do Arch",
	Executar:  instalarGo,
}

func instalarGo(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Go...")
	if err := s.Pacman(s.Dados.Lista("go", "pacman")...); err != nil {
		return err
	}
	s.UI.Sucesso("Go instalado.")
	return nil
}
