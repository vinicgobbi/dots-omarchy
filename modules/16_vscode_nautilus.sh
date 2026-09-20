#!/usr/bin/env bash

instalar_vscode_nautilus() {
    info "Instalando extensão do VSCode para o Nautilus..."

    pacman -S --needed --noconfirm nautilus-python python-gobject

    executar_como_usuario "
  rm -rf /tmp/vscode_nautilus
  git clone https://github.com/vinicgobbi/vscode_nautilus.git /tmp/vscode_nautilus
  bash /tmp/vscode_nautilus/install.sh
"
    sucesso "Extensão do VSCode para o Nautilus instalada."
}

registrar_modulo "vscode_nautilus" "Extensão VSCode no Nautilus" \
    "Adiciona 'Abrir com o VSCode' ao menu de contexto do Nautilus" \
    "instalar_vscode_nautilus" "pacotes_base"
