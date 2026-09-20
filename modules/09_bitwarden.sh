#!/usr/bin/env bash

instalar_bitwarden_nativo() {
    info "Instalando Bitwarden nativo (integração com navegador)..."
    instalar_pacotes_aur bitwarden-bin
    sucesso "Bitwarden instalado nativamente."
}

registrar_modulo "bitwarden" "Instalar Bitwarden nativo" \
    "Instala o Bitwarden desktop (necessário para integração nativa com o navegador)" \
    "instalar_bitwarden_nativo" "repositorios"
