#!/usr/bin/env bash

# Chave GPG oficial do Chaotic-AUR (aur.chaotic.cx) — repo binário que serve
# pacotes pré-compilados do AUR, evitando compilar tudo na mão.
CHAOTIC_AUR_KEY="3056513887B78AEB"
SUDOERS_AUR_FILE="/etc/sudoers.d/99-post-omarchy-aur"

# Habilita o repositório Chaotic-AUR seguindo o processo oficial
# (https://aur.chaotic.cx/): importa a chave, instala keyring+mirrorlist via
# pacman -U e inclui o mirrorlist no pacman.conf. Idempotente.
configurar_chaotic_aur() {
    if grep -q '^\[chaotic-aur\]' /etc/pacman.conf 2>/dev/null; then
        info "Chaotic-AUR já configurado."
        return
    fi

    info "Configurando o repositório binário Chaotic-AUR..."
    pacman-key --recv-key "$CHAOTIC_AUR_KEY" --keyserver keyserver.ubuntu.com
    pacman-key --lsign-key "$CHAOTIC_AUR_KEY"
    pacman -U --noconfirm \
        'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst' \
        'https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst'

    cat <<'EOT' >> /etc/pacman.conf

[chaotic-aur]
Include = /etc/pacman.d/chaotic-mirrorlist
EOT

    pacman -Sy
    sucesso "Chaotic-AUR configurado."
}

# O yay (e, por baixo, o makepkg) recusa rodar como root, mas o setup.sh
# inteiro roda via sudo. Sem essa liberação, todo módulo que precisa
# instalar algo da AUR pararia pedindo senha no meio da execução
# automática. Regra bem restrita (só o binário do pacman, não ALL) e
# temporária: removida em "Limpeza final" (modules/19_limpeza.sh).
permitir_sudo_pacman_temporario() {
    [[ -f "$SUDOERS_AUR_FILE" ]] && return

    info "Liberando sudo sem senha para o pacman (usuário $USER_NAME) — necessário para o yay instalar pacotes da AUR sem interação; revertido ao final em 'Limpeza final'."
    echo "$USER_NAME ALL=(ALL) NOPASSWD: /usr/bin/pacman" > "$SUDOERS_AUR_FILE"
    chmod 0440 "$SUDOERS_AUR_FILE"
    visudo -cf "$SUDOERS_AUR_FILE" || { rm -f "$SUDOERS_AUR_FILE"; erro "Falha ao validar a regra temporária de sudo para o pacman."; }
}

configurar_repositorios() {
    info "Configurando repositórios (Chaotic-AUR + yay)..."

    # O mirrorlist do Omarchy já aponta para o CDN próprio dele
    # (stable-mirror.omarchy.org), que também serve o repositório [omarchy];
    # por isso não rodamos o reflector aqui. O yay já vem pré-instalado.
    pacman -Sy --needed --noconfirm curl gnupg jq base-devel git

    configurar_chaotic_aur
    permitir_sudo_pacman_temporario

    comando_existe yay || erro "yay não encontrado — deveria vir pré-instalado no Omarchy."
    sucesso "Repositórios configurados."
}

registrar_modulo "repositorios" "Configurar repositórios" \
    "Habilita o Chaotic-AUR e libera o yay para instalar pacotes da AUR" \
    "configurar_repositorios"
