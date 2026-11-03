package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var atualizacao = &Modulo{
	ID:        "atualizacao",
	Grupo:     GrupoSistema,
	Titulo:    "Atualizar sistema",
	Descricao: "Roda o omarchy-update (snapshot, keyring, migrations e upgrade completo)",
	Executar:  atualizarSistema,
}

// comandoAtualizacao roda o omarchy-update sem nenhum terminal onde ele possa
// perguntar algo.
//
// Sozinho, ele se reexecuta dentro de `script`, que abre um pty novo: os passos
// passam a enxergar um terminal mesmo sem ninguém do outro lado, e o `-y` não
// cobre tudo. O `gum confirm` do omarchy-update-restart (kernel ou Hyprland
// atualizados) e o do omarchy-update-orphan-pkgs ficam esperando para sempre.
//
// Por isso o `script` é chamado aqui e OMARCHY_UPDATE_LOGGED=1 evita o dele. O
// log continua em /tmp/omarchy-update.log, que o omarchy-update-analyze-logs lê
// para detectar falha no initramfs. Dentro do pty, `setsid` tira o terminal de
// controle e o stdin vira /dev/null: um prompt inesperado (gum ou sudo) falha na
// hora em vez de travar, e cada passo segue o caminho "sem terminal" que o
// próprio Omarchy prevê.
const comandoAtualizacao = "script -qefc 'setsid -w env OMARCHY_UPDATE_LOGGED=1 omarchy-update -y </dev/null' " +
	"/tmp/omarchy-update.log </dev/null"

func atualizarSistema(s *sistema.Sistema) error {
	s.UI.Info("Atualizando pacotes do sistema...")
	// "omarchy-update" é a forma suportada: cria snapshot via Snapper antes de
	// atualizar, faz prune de cache, atualiza keyring, roda migrations do
	// Omarchy, atualiza AUR/mise e remove órfãos. Um "pacman -Syu" cru pula
	// tudo isso.
	//
	// Ele eleva privilégio internamente via "sudo pacman ..." e espera rodar
	// como o usuário logado, então roda como o usuário-alvo (o sudo não pede
	// senha: ver sistema.LiberarSudo). O log fixo em /tmp/omarchy-update.log,
	// se ficou de uma execução anterior com outro dono, é barrado pelo
	// fs.protected_regular do kernel até para o root — por isso removemos
	// antes (apagar o root sempre pode).
	if err := s.Remover("/tmp/omarchy-update.log"); err != nil {
		return err
	}
	if _, err := s.ComoUsuario(comandoAtualizacao, sistema.Opts{}); err != nil {
		return err
	}
	s.UI.Sucesso("Sistema atualizado.")
	return nil
}
