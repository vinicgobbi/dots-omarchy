from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo


def atualizar_sistema(s):
    ui.info("Atualizando pacotes do sistema...")
    # "omarchy-update" é a forma suportada: cria snapshot via Snapper antes de
    # atualizar, faz prune de cache, atualiza keyring, roda migrations do
    # Omarchy, atualiza AUR/mise e remove órfãos. Um "pacman -Syu" cru pula
    # tudo isso.
    #
    # Ele eleva privilégio internamente via "sudo pacman ..." e espera rodar
    # como o usuário logado, então roda como o usuário-alvo (pode pedir a
    # senha de sudo dele no meio). Também grava um log fixo em
    # /tmp/omarchy-update.log: se esse arquivo ficou de uma execução anterior
    # com outro dono, o fs.protected_regular do kernel barra até o root de
    # reabri-lo — por isso removemos antes (apagar o root sempre pode).
    s.remover(Path("/tmp/omarchy-update.log"))
    s.como_usuario("omarchy-update -y")
    ui.sucesso("Sistema atualizado.")


MODULO = Modulo("atualizacao", "Atualizar sistema",
                "Roda o omarchy-update (snapshot, keyring, migrations e upgrade completo)",
                atualizar_sistema)
