#!/usr/bin/env bash

importar_ovpn() {
    info "Importando perfis OpenVPN para o NetworkManager..."

    pacman -Qq networkmanager-openvpn &>/dev/null || pacman -S --needed --noconfirm networkmanager-openvpn

    local ovpn_dir
    ovpn_dir="$SCRIPT_DIR/OVPN"

    if [[ ! -d "$ovpn_dir" ]] || ! compgen -G "$ovpn_dir/*.ovpn" > /dev/null; then
        aviso "Nenhum arquivo .ovpn encontrado em $ovpn_dir — pulando importação."
        return
    fi

    local arquivo nome dns dominios
    for arquivo in "$ovpn_dir"/*.ovpn; do
        nome="$(basename "$arquivo" .ovpn)"

        # Reimportar do zero para o script continuar idempotente em reexecuções.
        if nmcli -g NAME connection show | grep -Fxq "$nome"; then
            nmcli connection delete "$nome" > /dev/null
        fi

        if ! nmcli connection import type openvpn file "$arquivo" > /dev/null; then
            aviso "Falha ao importar $arquivo, pulando."
            continue
        fi

        # DNS e domínio de busca não vêm sempre preenchidos pela importação
        # automática do plugin; quando o .ovpn declara essas diretivas, aplica
        # manualmente via nmcli.
        dns=$(grep -oP '^\s*dhcp-option\s+DNS\s+\K\S+' "$arquivo" | paste -sd ' ' -)
        dominios=$(grep -oP '^\s*dhcp-option\s+DOMAIN(-SEARCH)?\s+\K\S+' "$arquivo" | paste -sd ' ' -)

        [[ -n "$dns" ]] && nmcli connection modify "$nome" ipv4.dns "$dns"
        [[ -n "$dominios" ]] && nmcli connection modify "$nome" ipv4.dns-search "$dominios"

        sucesso "Perfil '$nome' importado."
    done
}

registrar_modulo "ovpn" "Importar perfis OpenVPN" \
    "Instala o plugin OpenVPN do NetworkManager e importa os .ovpn de ./OVPN (aplicando DNS/domínio quando presentes no arquivo)" \
    "importar_ovpn"
