package modulos

import (
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var webapps = &Modulo{
	ID:     "webapps",
	Titulo: "Criar webapps",
	Descricao: "Apaga os webapps de fábrica do Omarchy e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, " +
		"GitHub, Claude, GLPI e o admin console do Tailscale",
	Executar: criarWebapps,
}

func criarWebapps(s *sistema.Sistema) error {
	s.UI.Info("Removendo os webapps de fábrica do Omarchy (omarchy-webapp-remove-all)...")
	ok, err := s.Tentar("omarchy-webapp-remove-all", sistema.Opts{})
	if err != nil {
		return err
	}
	if !ok {
		s.UI.Aviso("Falha ao remover os webapps do Omarchy.")
	}

	for _, w := range s.Dados.Webapps {
		// Ícone local (ex.: assets/icons/glpi.svg) é relativo à raiz do projeto.
		icone := w.Icone
		if !strings.Contains(icone, "://") {
			icone = filepath.Join(s.Raiz, icone)
		}
		ok, err := s.Tentar(sistema.Juntar("omarchy-webapp-install", w.Nome, w.URL, icone), sistema.Opts{Quieto: true})
		if err != nil {
			return err
		}
		if ok {
			s.UI.Info("Webapp criado: " + w.Nome)
		} else {
			s.UI.Aviso("Não foi possível criar o webapp " + w.Nome + ".")
		}
	}
	s.UI.Sucesso("Webapps configurados.")
	return nil
}
