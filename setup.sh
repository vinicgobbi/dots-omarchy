#!/usr/bin/env bash
# Mantém o "sudo ./setup.sh" de sempre; a lógica toda está em setup.py.
exec python3 "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/setup.py" "$@"
