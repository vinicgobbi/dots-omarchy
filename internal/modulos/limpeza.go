package modulos

import (
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var limpeza = &Modulo{
	ID:        "limpeza",
	Grupo:     GrupoFinalizacao,
	Titulo:    "Limpeza final",
	Descricao: "Remove pacotes órfãos e limpa o cache do pacman",
	Executar:  limpezaFinal,
}

func limpezaFinal(s *sistema.Sistema) error {
	s.UI.Info("Limpando o sistema...")
	res, err := s.Rodar([]string{"pacman", "-Qtdq"}, sistema.Opts{SemCheck: true, Capturar: true, Leitura: true})
	if err != nil {
		return err
	}
	if orfaos := strings.Fields(res.Saida); len(orfaos) > 0 {
		if err := s.Run(append([]string{"pacman", "-Rns", "--noconfirm"}, orfaos...)...); err != nil {
			return err
		}
	}
	if err := s.Run("pacman", "-Sc", "--noconfirm"); err != nil {
		return err
	}
	s.UI.Sucesso("Limpeza concluída.")
	return nil
}
