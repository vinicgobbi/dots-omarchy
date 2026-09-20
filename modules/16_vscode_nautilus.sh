#!/usr/bin/env bash

instalar_vscode_nautilus() {
    info "Instalando extensões do VSCode e de copiar caminho para o Nautilus..."

    pacman -S --needed --noconfirm nautilus-python python-gobject wl-clipboard xclip

    executar_como_usuario "
  rm -rf /tmp/vscode_nautilus
  git clone https://github.com/vinicgobbi/vscode_nautilus.git /tmp/vscode_nautilus
  bash /tmp/vscode_nautilus/install.sh

  rm -rf /tmp/nautilus-copypath
  git clone https://github.com/vinicgobbi/nautilus-copypath.git /tmp/nautilus-copypath
  bash /tmp/nautilus-copypath/install.sh
"
    sucesso "Extensões do VSCode e de copiar caminho instaladas no Nautilus."
}

registrar_modulo "vscode_nautilus" "Extensões do Nautilus (VSCode e copiar caminho)" \
    "Adiciona 'Abrir com o VSCode' e 'Copiar caminho' ao menu de contexto do Nautilus" \
    "instalar_vscode_nautilus" "pacotes_base"
