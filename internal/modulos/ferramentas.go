package modulos

import (
	"errors"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var ferramentas = &Modulo{
	ID:        "ferramentas",
	Grupo:     GrupoTerminal,
	Titulo:    "Ferramentas de terminal",
	Descricao: "git, curl, jq, bat, gh, tea, nmap, pipx, ffmpeg, bashtop e FUSE 2 para AppImages",
	Executar:  instalarFerramentas,
}

// instalarFerramentas tenta tudo numa chamada só; se algum pacote faltar, instala
// um por um e só avisa dos que falharem, sem derrubar o módulo.
func instalarFerramentas(s *sistema.Sistema) error {
	s.UI.Info("Instalando ferramentas de terminal...")
	pacotes := s.Dados.Lista("ferramentas", "pacman")
	err := s.Pacman(pacotes...)
	var falha *sistema.FalhaComando
	if err != nil && !errors.As(err, &falha) {
		return err
	}
	if err != nil {
		for _, p := range pacotes {
			if err := s.Pacman(p); errors.As(err, &falha) {
				s.UI.Aviso("Não foi possível instalar " + p + ": " + err.Error())
			} else if err != nil {
				return err
			}
		}
	}
	s.UI.Sucesso("Ferramentas de terminal instaladas.")
	return nil
}
