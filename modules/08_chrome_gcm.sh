#!/usr/bin/env bash

instalar_chrome_e_gcm() {
    info "Instalando Git Credential Manager e Google Chrome..."
    GCM_URL=$(curl -sL https://api.github.com/repos/git-ecosystem/git-credential-manager/releases/latest | jq -r '.assets[] | select(.name | endswith(".tar.gz") and contains("linux-x64") and (contains("symbols") | not)) | .browser_download_url' | head -n 1)

    # GCM (tar.gz local) e Chrome são independentes: o download do GCM roda
    # em segundo plano enquanto o Chrome é instalado.
    curl -sSL -o /tmp/gcm.tar.gz "$GCM_URL" &
    pid_gcm=$!

    # omarchy-install-browser instala via yay e ainda configura política/tema
    # do Chrome — a política grava em /etc/opt/chrome/policies/managed, o que
    # pode pedir a senha de sudo do usuário-alvo no meio da instalação. Roda
    # em primeiro plano de propósito, pra não perder esse prompt.
    executar_como_usuario "omarchy-install-browser chrome"
    wait "$pid_gcm"

    mkdir -p /usr/local/gcm
    tar -xzf /tmp/gcm.tar.gz -C /usr/local/gcm
    ln -sf /usr/local/gcm/git-credential-manager /usr/local/bin/git-credential-manager
    sucesso "Chrome e GCM instalados."
}

registrar_modulo "chrome_gcm" "Chrome + Git Credential Manager" \
    "Instala o Google Chrome (integrado ao tema do Omarchy) e o Git Credential Manager" \
    "instalar_chrome_e_gcm"
