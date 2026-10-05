#!/usr/bin/env bash

# Temas extras vêm do GitHub de cada autor e podem trazer hooks (theme-set.d)
# que rodam como o usuário a cada troca de tema; por isso, como os plugins,
# só instala depois de confirmar. Sem terminal interativo, não instala nada.
instalar_temas_omarchy() {
    if [[ "${#OMARCHY_TEMAS[@]}" -eq 0 ]]; then
        aviso "Nenhum tema listado em config.sh; nada a fazer."
        return
    fi

    if [[ ! -t 0 ]]; then
        aviso "Sem terminal interativo para confirmar: temas de terceiros NÃO foram instalados."
        return
    fi

    echo
    echo "Temas que serão instalados (código de terceiros, podem trazer hooks"
    echo "que rodam a cada troca de tema):"
    local url
    for url in "${OMARCHY_TEMAS[@]}"; do
        echo "  - $url"
    done
    echo
    if ! confirmar "Instalar os ${#OMARCHY_TEMAS[@]} temas acima?"; then
        aviso "Temas não instalados (escolha do usuário)."
        return
    fi

    # O omarchy-theme-install clona e já aplica o tema, então o último da
    # lista é o que fica ativo. Fora da sessão gráfica a aplicação pode
    # falhar, mas o clone fica feito e o tema aparece no seletor.
    for url in "${OMARCHY_TEMAS[@]}"; do
        info "Tema do Omarchy: $url"
        executar_como_usuario "omarchy-theme-install '$url'" \
            || aviso "Não foi possível instalar/aplicar o tema $url; rode depois: omarchy-theme-install $url"
    done

    sucesso "Temas instalados. Troque pelo seletor de temas do Omarchy, se quiser."
}

registrar_modulo "temas_omarchy" "Instalar temas extras" \
    "Pergunta antes de instalar os temas listados em config.sh (o último fica ativo)" \
    "instalar_temas_omarchy"
