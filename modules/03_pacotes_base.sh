#!/usr/bin/env bash

instalar_pacotes_base() {
    info "Instalando pacotes base e utilitários..."
    # Docker, PHP/composer e unixODBC são oficiais (repo extra/core). O
    # repositório [omarchy] já traz o visual-studio-code-bin pré-compilado
    # (mesmo pacote que "omarchy install editor vscode" usa); msodbcsql e
    # mssql-tools só existem na AUR (resolvidos pelo Chaotic-AUR, sem compilar).
    pacman -S --needed --noconfirm flatpak zsh git curl jq xdg-user-dirs \
        docker docker-compose docker-buildx php composer unixodbc bat \
        visual-studio-code-bin
    instalar_pacotes_aur msodbcsql mssql-tools
    aplicar_config_omarchy_vscode

    # O pacote da AUR instala em /opt/mssql-tools (sem versão).
    echo 'export PATH="$PATH:/opt/mssql-tools/bin"' > /etc/profile.d/mssql-tools.sh
    sucesso "Pacotes e MS SQL instalados."
}

# Replica o que "omarchy-install-editor-vscode" faz para o usuário-alvo
# (senha via gnome-libsecret, autoupdate desligado — o Omarchy já atualiza o
# VSCode via pacman — e o tema atual do Omarchy aplicado no editor), sem o
# lançamento automático da GUI ao final, que assume uma sessão gráfica
# interativa e não faz sentido no meio de uma instalação em lote.
aplicar_config_omarchy_vscode() {
    executar_como_usuario "
  mkdir -p \$HOME/.vscode \$HOME/.config/Code/User
  printf '{\n  \"password-store\":\"gnome-libsecret\"\n}\n' > \$HOME/.vscode/argv.json
  printf '{\n  \"update.mode\": \"none\"\n}\n' > \$HOME/.config/Code/User/settings.json
  omarchy-theme-set-vscode
"
}

registrar_modulo "pacotes_base" "Instalar pacotes base" \
    "Docker, VSCode, PHP, Zsh e ferramentas de SQL Server" \
    "instalar_pacotes_base" "repositorios"
