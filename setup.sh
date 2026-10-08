#!/usr/bin/env bash
# Mantém o "sudo ./setup.sh" de sempre. No clone do repositório, compila o
# post-omarchy (Go) quando o binário falta ou ficou mais velho que o código; no
# zip do release, o binário já vem pronto. A lógica toda está em main.go e
# internal/.
set -euo pipefail

raiz="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
binario="$raiz/post-omarchy"

desatualizado() {
  # Pacote do release: só o binário, sem código para compilar.
  [[ ! -f $raiz/main.go ]] && return 1
  [[ ! -x $binario ]] && return 0
  [[ -n $(find "$raiz/main.go" "$raiz/go.mod" "$raiz/go.sum" "$raiz/internal" \
    -newer "$binario" -print -quit) ]]
}

if desatualizado; then
  if ! command -v go >/dev/null; then
    if [[ $EUID -ne 0 ]]; then
      echo "[-] O Go não está instalado. Rode com sudo (o setup instala o pacote 'go') ou: sudo pacman -S go" >&2
      exit 1
    fi
    echo "[*] Instalando o Go para compilar o setup..."
    pacman -S --needed --noconfirm go
  fi
  echo "[*] Compilando o post-omarchy..."
  (cd "$raiz" && go build -trimpath -ldflags='-s -w' -o "$binario" .)
  # Compilado via sudo: devolve o binário ao dono do projeto.
  chown --reference="$raiz" "$binario" 2>/dev/null || true
fi

exec "$binario" "$@"
