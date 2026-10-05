#!/usr/bin/env bash

instalar_chrome_e_gcm() {
    info "Instalando Git Credential Manager e Google Chrome..."
    GCM_URL=$(curl -sL https://api.github.com/repos/git-ecosystem/git-credential-manager/releases/latest | jq -r '.assets[] | select(.name | endswith(".tar.gz") and contains("linux-x64") and (contains("symbols") | not)) | .browser_download_url' | head -n 1)

    # GCM (tar.gz local) e Chrome são independentes: o download do GCM roda
    # em segundo plano enquanto o Chrome é instalado.
    curl -sSL -o /tmp/gcm.tar.gz "$GCM_URL" &
    pid_gcm=$!

    # omarchy-install-browser instala via yay e tenta criar a pasta de política
    # do Chrome com sudo. Sob `su -c` o sudo não consegue pedir a senha
    # ("conversation failed") e o script segue sem a pasta, então o tema nunca
    # chega ao Chrome. Como aqui já somos root, criamos a pasta com o próprio
    # helper do Omarchy e reaplicamos o tema como o usuário.
    executar_como_usuario "omarchy-install-browser chrome"
    source "${OMARCHY_PATH:-/usr/share/omarchy}/install/helpers/browser-policy.sh"
    browser_policy_setup_dir /etc/opt/chrome/policies/managed
    executar_como_usuario "omarchy-theme-set-browser"
    wait "$pid_gcm"

    mkdir -p /usr/local/gcm
    tar -xzf /tmp/gcm.tar.gz -C /usr/local/gcm
    ln -sf /usr/local/gcm/git-credential-manager /usr/local/bin/git-credential-manager
    sucesso "Chrome e GCM instalados."
}

registrar_modulo "chrome_gcm" "Chrome + Git Credential Manager" \
    "Instala o Google Chrome (integrado ao tema do Omarchy) e o Git Credential Manager" \
    "instalar_chrome_e_gcm"
