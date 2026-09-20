#!/usr/bin/env bash

instalar_rust_tools() {
    info "Instalando rustup e compilando eza/topgrade via cargo..."

    pacman -S --needed --noconfirm base-devel

    executar_como_usuario "
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --default-toolchain stable
  source \"\$HOME/.cargo/env\"

  cargo install --locked eza topgrade
"
    sucesso "rustup, eza e topgrade instalados para $USER_NAME."
}

registrar_modulo "rust_tools" "Instalar Rust, eza e topgrade" \
    "rustup e compilação de eza e topgrade via cargo" \
    "instalar_rust_tools"
