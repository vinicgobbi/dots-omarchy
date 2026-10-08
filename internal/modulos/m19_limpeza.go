package modulos

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var limpeza = &Modulo{
	ID:        "limpeza",
	Titulo:    "Limpeza final",
	Descricao: "Remove pacotes órfãos, limpa o cache do pacman e remove o sudo temporário do yay",
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

	if _, err := os.Stat(sudoersAURFile); err == nil {
		if err := s.Remover(sudoersAURFile); err != nil {
			return err
		}
		s.UI.Info("Regra temporária de sudo sem senha para o pacman (liberada para o yay) removida.")
	}

	// Mesmo alcance do "rm -rf /tmp/*": não mexe em entradas ocultas.
	itens, _ := filepath.Glob("/tmp/*")
	for _, item := range itens {
		s.Remover(item)
	}

	s.UI.Sucesso("Instalação finalizada com sucesso!")
	s.UI.Info("Recomenda-se reiniciar a máquina para aplicar as mudanças de grupo e kernel.")
	return nil
}
