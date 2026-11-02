package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var docker = &Modulo{
	ID:        "docker",
	Grupo:     GrupoDesenvolvimento,
	Titulo:    "Docker",
	Descricao: "Docker, Buildx e Compose; o serviço fica ativo e o usuário entra no grupo docker",
	Executar:  instalarDocker,
}

func instalarDocker(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Docker...")
	if err := s.Pacman(s.Dados.Lista("docker", "pacman")...); err != nil {
		return err
	}
	if err := s.Run("systemctl", "enable", "--now", "docker"); err != nil {
		return err
	}
	if err := s.Run("usermod", "-aG", "docker", s.Usuario()); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("Docker instalado. '%s' foi adicionado ao grupo docker "+
		"(é preciso reabrir a sessão para valer).", s.Usuario()))
	return nil
}
