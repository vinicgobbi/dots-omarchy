package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var atualizacao = &Modulo{
	ID:         "atualizacao",
	Titulo:     "Atualizar sistema",
	Descricao:  "Roda o omarchy-update (snapshot, keyring, migrations e upgrade completo)",
	Executar:   atualizarSistema,
	Interativo: true,
}

func atualizarSistema(s *sistema.Sistema) error {
	s.UI.Info("Atualizando pacotes do sistema...")
	// "omarchy-update" é a forma suportada: cria snapshot via Snapper antes de
	// atualizar, faz prune de cache, atualiza keyring, roda migrations do
	// Omarchy, atualiza AUR/mise e remove órfãos. Um "pacman -Syu" cru pula
	// tudo isso.
	//
	// Ele eleva privilégio internamente via "sudo pacman ..." e espera rodar
	// como o usuário logado, então roda como o usuário-alvo e com o terminal
	// (pode pedir a senha de sudo dele no meio). Também grava um log fixo em
	// /tmp/omarchy-update.log: se esse arquivo ficou de uma execução anterior
	// com outro dono, o fs.protected_regular do kernel barra até o root de
	// reabri-lo — por isso removemos antes (apagar o root sempre pode).
	if err := s.Remover("/tmp/omarchy-update.log"); err != nil {
		return err
	}
	if _, err := s.ComoUsuario("omarchy-update -y", sistema.Opts{Interativo: true}); err != nil {
		return err
	}
	s.UI.Sucesso("Sistema atualizado.")
	return nil
}
