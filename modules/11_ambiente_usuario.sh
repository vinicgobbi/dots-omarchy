#!/usr/bin/env bash

configurar_usuario() {
    info "Aplicando configurações locais para $USER_NAME..."
    usermod -aG docker "$USER_NAME"

    # chsh só aceita contas presentes de fato em /etc/passwd; contas de
    # domínio (AD/LDAP via sssd/winbind) resolvem por NSS mas não estão no
    # arquivo, e o chsh recusa. Comparamos a resolução restrita à fonte
    # "files" (local) para distinguir os dois casos.
    if getent -s files passwd "$USER_NAME" &>/dev/null; then
        chsh -s "$(command -v zsh)" "$USER_NAME"
    else
        aviso "'$USER_NAME' é uma conta de domínio (AD/LDAP), não local: pulando 'chsh' (não dá para trocar o shell padrão por aqui nesse caso)."
    fi

    # O tema (GTK, ícones, Nautilus etc.) já é gerido pelo próprio Omarchy
    # ("omarchy theme"), então não aplicamos gsettings/adw-gtk3/Yaru aqui.
    executar_como_usuario "
  # Git Credential Manager
  git-credential-manager configure
  git config --global credential.credentialStore secretservice

  # FNM (download do binário, sem tocar no shell ainda) e o clone+bootstrap
  # dos dotfiles (Oh My Zsh, plugins, tema, fontes e config do Solaar) não
  # dependem um do outro: rodam em paralelo em vez de em série.
  ( curl -fsSL https://fnm.vercel.app/install | bash -s -- --skip-shell ) &
  pid_fnm=\$!
  ( rm -rf /tmp/dotfiles && git clone https://github.com/vinicgobbi/Dotfiles.git /tmp/dotfiles && bash /tmp/dotfiles/bootstrap.sh ) &
  pid_dotfiles=\$!
  wait \"\$pid_fnm\" \"\$pid_dotfiles\"

  # Node LTS via FNM (agora que o binário já foi baixado acima)
  export PATH=\"\$HOME/.local/share/fnm:\$PATH\"
  eval \"\$(fnm env --shell zsh)\"

  fnm install --lts
  fnm default \$(fnm current)

  # Injeta a inicialização do fnm explicitamente no .zshrc (precisa rodar
  # depois do Oh My Zsh, que é quem cria/substitui o ~/.zshrc)
  echo 'export PATH=\"\$HOME/.local/share/fnm:\$PATH\"' >> ~/.zshrc
  echo 'eval \"\$(fnm env --shell zsh)\"' >> ~/.zshrc

  # Diretório de Projetos (XDG e Bookmarks)
  if [[ \"\$LANG\" == pt_* ]]; then
      DIR_NAME=\"Projetos\"
  else
      DIR_NAME=\"Projects\"
  fi
  PROJECTS_DIR=\"\$HOME/\$DIR_NAME\"
  mkdir -p \"\$PROJECTS_DIR\"

  xdg-user-dirs-update --set PROJECTS \"\$PROJECTS_DIR\"

  BOOKMARKS_FILE=\"\$HOME/.config/gtk-3.0/bookmarks\"
  mkdir -p \"\$(dirname \"\$BOOKMARKS_FILE\")\"
  touch \"\$BOOKMARKS_FILE\"
  if ! grep -q \"file://\$PROJECTS_DIR\" \"\$BOOKMARKS_FILE\"; then
      echo \"file://\$PROJECTS_DIR\" >> \"\$BOOKMARKS_FILE\"
  fi

  # Autostart do Solaar (prioriza o .desktop do pacote nativo; usa o do Flatpak como alternativa)
  mkdir -p \"\$HOME/.config/autostart\"
  cp /usr/share/applications/solaar.desktop \"\$HOME/.config/autostart/\" 2>/dev/null || \
      cp /var/lib/flatpak/exports/share/applications/io.github.pwr_solaar.solaar.desktop \"\$HOME/.config/autostart/\" 2>/dev/null || true
"
    sucesso "Ambiente de usuário configurado."
}

registrar_modulo "ambiente_usuario" "Configurar ambiente do usuário" \
    "Zsh, Oh My Zsh, fnm/Node e dotfiles de shell (o tema visual fica com o Omarchy)" \
    "configurar_usuario" "pacotes_base"
