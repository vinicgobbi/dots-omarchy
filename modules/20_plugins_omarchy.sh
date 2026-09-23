#!/usr/bin/env bash

# Plugins são código de terceiros que roda dentro do omarchy-shell: por isso
# este é o último módulo e só instala depois de mostrar os riscos e o usuário
# confirmar explicitamente. Sem terminal interativo, não instala nada.
avisar_riscos_plugins() {
    echo
    echo -e "${NEGRITO}${VERMELHO}ATENÇÃO: plugins de terceiros${NC}"
    echo -e "Os plugins abaixo são baixados direto do GitHub de cada autor e ${NEGRITO}não${NC}"
    echo "são revisados nem mantidos pelo Omarchy nem por este script."
    echo
    echo "  - Rodam como código arbitrário, sem sandbox, dentro do omarchy-shell,"
    echo "    com as mesmas permissões do seu usuário: podem ler e alterar seus"
    echo "    arquivos, executar comandos e acessar a rede."
    echo "  - É instalada a versão atual de cada repositório, que pode ter mudado"
    echo "    (ou sido comprometida) desde que esta lista foi montada."
    echo "  - Um plugin com bug pode travar ou deixar lenta a barra e o shell."
    echo "  - Alguns pedem passos extras (dependências, scripts de instalação);"
    echo "    veja o README de cada um."
    echo
    echo "Revise o código antes, se puder. Plugins que serão instalados:"
    local url
    for url in "${OMARCHY_PLUGINS[@]}"; do
        echo "  - $url"
    done
    echo
}

instalar_plugins_omarchy() {
    if [[ "${#OMARCHY_PLUGINS[@]}" -eq 0 ]]; then
        aviso "Nenhum plugin listado em config.sh; nada a fazer."
        return
    fi

    if [[ ! -t 0 ]]; then
        aviso "Sem terminal interativo para confirmar: plugins de terceiros NÃO foram instalados."
        return
    fi

    avisar_riscos_plugins
    if ! confirmar "Entendi os riscos. Instalar os ${#OMARCHY_PLUGINS[@]} plugins acima?"; then
        aviso "Plugins de terceiros não instalados (escolha do usuário)."
        return
    fi

    # --yes porque a confirmação já foi feita acima (e o su não tem o gum
    # interativo do omarchy). Cada plugin é best-effort: precisa de rede e
    # falha se já estiver instalado.
    local url
    for url in "${OMARCHY_PLUGINS[@]}"; do
        info "Plugin do Omarchy: $url"
        executar_como_usuario "omarchy plugin add '$url' --yes" \
            || aviso "Não foi possível instalar o plugin $url; rode depois: omarchy plugin add $url"
    done

    sucesso "Plugins instalados. A posição deles na barra vem do shell.json dos dots."
}

registrar_modulo "plugins_omarchy" "Instalar plugins de terceiros" \
    "Mostra os riscos e pergunta antes de instalar os plugins da barra listados em config.sh" \
    "instalar_plugins_omarchy"
