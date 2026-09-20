#!/usr/bin/env bash
# ==========================================
# Utilitários gerais
# ==========================================

comando_existe() {
    command -v "$1" &>/dev/null
}

# Registro de módulos: cada arquivo em modules/ chama isto no final para
# se anunciar ao orquestrador (usado para montar o menu e a ordem de execução).
MOD_IDS=()
MOD_TITLES=()
MOD_DESCS=()
MOD_FUNCS=()
MOD_DEPS=()

# Uso: registrar_modulo <id> <titulo> <descricao> <funcao> [id_dependencia...]
# Os ids de dependência (opcionais) são de outros módulos já registrados com
# registrar_modulo — ver resolver_dependencias_modulos, chamada pelo setup.sh
# depois do menu_checklist para garantir que eles também sejam selecionados.
registrar_modulo() {
    MOD_IDS+=("$1")
    MOD_TITLES+=("$2")
    MOD_DESCS+=("$3")
    MOD_FUNCS+=("$4")
    shift 4
    MOD_DEPS+=("$*")
}

# Dado o resultado do menu_checklist, marca como selecionado (1) qualquer
# módulo que seja dependência de um módulo já selecionado, mesmo que o
# usuário o tenha desmarcado — e avisa qual módulo puxou qual. Sem isso, dá
# pra selecionar só "Extensões PHP (SQL Server)" sem "Configurar
# repositórios"/"Instalar pacotes base" e a compilação do sqlsrv falha na
# hora, porque os headers do driver ODBC nunca foram instalados.
# Repete em ponto fixo para cobrir cadeias de dependência (A depende de B,
# que depende de C).
resolver_dependencias_modulos() {
    local -n _ids="$1" _deps="$2" _titulos="$3" _sel="$4"
    local n="${#_ids[@]}" mudou=1 i j dep_id idx

    while [[ "$mudou" -eq 1 ]]; do
        mudou=0
        for ((i = 0; i < n; i++)); do
            [[ "${_sel[i]}" == "1" && -n "${_deps[i]}" ]] || continue
            for dep_id in ${_deps[i]}; do
                idx=-1
                for ((j = 0; j < n; j++)); do
                    [[ "${_ids[j]}" == "$dep_id" ]] && idx=$j && break
                done
                if [[ "$idx" -ge 0 && "${_sel[idx]}" != "1" ]]; then
                    _sel[idx]=1
                    mudou=1
                    aviso "'${_titulos[i]}' depende de '${_titulos[idx]}': selecionando automaticamente."
                fi
            done
        done
    done
}

# Executa um bloco de shell como o usuário alvo (login shell, com $HOME correto).
executar_como_usuario() {
    su - "$USER_NAME" -c "$1"
}

# Instala pacotes via yay (só faz sentido na família "arch") como o usuário
# alvo — o makepkg recusa rodar como root. Resolve tanto pacotes binários do
# Chaotic-AUR (instala direto, sem compilar) quanto pacotes que só existem na
# AUR de fato (compila). Depende do módulo "repositorios" ter liberado sudo
# sem senha para o pacman nesse usuário antes (ver modules/02_repositorios.sh).
instalar_pacotes_aur() {
    executar_como_usuario "yay -S --needed --noconfirm $*"
}

# Baixa uma URL com algumas tentativas antes de desistir, para tolerar
# instabilidade de rede em downloads de repositórios/pacotes externos.
download_com_retry() {
    local url="$1" destino="$2" tentativas=3 i

    for ((i = 1; i <= tentativas; i++)); do
        if curl -fsSL --retry 2 -o "$destino" "$url"; then
            return 0
        fi
        aviso "Falha ao baixar $url (tentativa $i/$tentativas)."
    done
    erro "Não foi possível baixar $url após $tentativas tentativas."
}
