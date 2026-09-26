#!/usr/bin/env bash

instalar_vscode_nautilus() {
    info "Instalando extensões do VSCode, de copiar caminho e de abrir no terminal para o Nautilus..."

    pacman -S --needed --noconfirm nautilus-python python-gobject wl-clipboard xclip

    executar_como_usuario "
  rm -rf /tmp/vscode_nautilus
  git clone https://github.com/vinicgobbi/vscode_nautilus.git /tmp/vscode_nautilus
  bash /tmp/vscode_nautilus/install.sh

  rm -rf /tmp/nautilus-copypath
  git clone https://github.com/vinicgobbi/nautilus-copypath.git /tmp/nautilus-copypath
  bash /tmp/nautilus-copypath/install.sh

  rm -rf /tmp/omarchy-nautilus-openinterminal
  git clone https://github.com/vinicgobbi/omarchy-nautilus-openinterminal.git /tmp/omarchy-nautilus-openinterminal
  bash /tmp/omarchy-nautilus-openinterminal/install.sh
"
    sucesso "Extensões do VSCode, de copiar caminho e de abrir no terminal instaladas no Nautilus."
}

registrar_modulo "vscode_nautilus" "Extensões do Nautilus (VSCode, copiar caminho e abrir no terminal)" \
    "Adiciona 'Abrir com o VSCode', 'Copiar caminho' e 'Abrir no Terminal' ao menu de contexto do Nautilus" \
    "instalar_vscode_nautilus" "pacotes_base"
