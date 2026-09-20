#!/usr/bin/env bash

atualizar_sistema() {
    info "Atualizando pacotes do sistema..."
    # "omarchy-update" é a forma suportada: cria snapshot via Snapper antes de
    # atualizar, faz prune de cache, atualiza keyring, roda migrations do
    # Omarchy, atualiza AUR/mise e remove órfãos. Um "pacman -Syu" cru pula
    # tudo isso.
    #
    # O script internamente eleva privilégio via "sudo pacman ..." e espera
    # rodar como o usuário logado (não como root direto) — roda como
    # USER_NAME, igual aos outros "omarchy-install-*" deste projeto (pode
    # pedir a senha de sudo do usuário-alvo no meio do processo). Ele também
    # grava um log fixo em /tmp/omarchy-update.log: se esse arquivo ficou de
    # uma execução anterior com outro dono, o hardening padrão do kernel
    # (fs.protected_regular) barra até o root de reabri-lo — por isso
    # removemos antes, já que apagar (diferente de reabrir) o root sempre pode.
    rm -f /tmp/omarchy-update.log
    executar_como_usuario "omarchy-update -y"
    sucesso "Sistema atualizado."
}

registrar_modulo "atualizacao" "Atualizar sistema" \
    "Roda o omarchy-update (snapshot, keyring, migrations e upgrade completo)" \
    "atualizar_sistema"
