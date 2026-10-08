package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var tailscale = &Modulo{
	ID:        "tailscale",
	Titulo:    "Instalar Tailscale",
	Descricao: "Tailscale (VPN mesh) com widget na barra e webapp do admin console",
	Executar:  instalarTailscale,
}

func instalarTailscale(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Tailscale...")

	// tailscale é pacote oficial do repo "extra" — mesmo pacote que
	// "omarchy install service tailscale" usa.
	if err := s.Pacman(s.Dados.Lista("tailscale", "pacman")...); err != nil {
		return err
	}
	if err := s.Run("systemctl", "enable", "--now", "tailscaled"); err != nil {
		return err
	}
	if err := s.Run("tailscale", "set", "--operator="+s.Usuario()); err != nil {
		return err
	}
	if err := integrarTailscaleOmarchy(s); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("Tailscale instalado e '%s' definido como operator.", s.Usuario()))
	return nil
}

// integrarTailscaleOmarchy replica o que "omarchy-install-service-tailscale"
// faz: widget na barra, atalho de webapp pro admin console e recebimento de
// Taildrop em ~/Downloads. Cada etapa é best-effort (não aborta o módulo): a
// sessão gráfica do usuário-alvo pode não estar ativa ainda nesse momento.
func integrarTailscaleOmarchy(s *sistema.Sistema) error {
	quieto := sistema.Opts{Quieto: true}

	ok, err := s.Tentar(fmt.Sprintf("XDG_RUNTIME_DIR='/run/user/%d' systemctl --user enable --now "+
		"omarchy-tailscale-receive.service", s.Conta.UID), quieto)
	if err != nil {
		return err
	}
	if ok {
		s.UI.Info("Recebimento de Taildrop habilitado em ~/Downloads.")
	} else {
		s.UI.Aviso(fmt.Sprintf("Não foi possível habilitar o recebimento de Taildrop agora (sessão gráfica de "+
			"'%s' inativa?); rode depois de logar: "+
			"systemctl --user enable --now omarchy-tailscale-receive.service", s.Usuario()))
	}

	if ok, err = s.Tentar("omarchy-plugin-enable omarchy.tailscale", quieto); err != nil {
		return err
	}
	if ok {
		s.UI.Info("Tailscale adicionado à barra do Omarchy.")
	} else {
		s.UI.Aviso("Não foi possível adicionar o Tailscale à barra do Omarchy agora.")
	}

	return criarWebappTailscaleAdmin(s)
}

// criarWebappTailscaleAdmin: o módulo "webapps" também cria este atalho (a
// entrada vem do mesmo data/webapps.json), porque o omarchy-webapp-remove-all
// apaga todos os webapps de uma vez.
func criarWebappTailscaleAdmin(s *sistema.Sistema) error {
	for _, w := range s.Dados.Webapps {
		if w.Nome != "Tailscale" {
			continue
		}
		ok, err := s.Tentar(sistema.Juntar("omarchy-webapp-install", w.Nome, w.URL, w.Icone),
			sistema.Opts{Quieto: true})
		if err != nil {
			return err
		}
		if !ok {
			s.UI.Aviso("Não foi possível criar o atalho de webapp do Tailscale Admin Console.")
		}
	}
	return nil
}
