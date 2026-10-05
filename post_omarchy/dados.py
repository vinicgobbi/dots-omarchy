"""Listas editáveis do projeto, em data/*.json."""

import json
from pathlib import Path

ARQUIVOS = ("flatpaks", "plugins", "themes", "webapps", "packages", "dots")


def carregar(raiz: Path) -> dict:
    """Lê todos os JSON de data/. Um JSON inválido para o setup logo no início,
    antes de qualquer mudança no sistema."""
    return {nome: json.loads((raiz / "data" / f"{nome}.json").read_text()) for nome in ARQUIVOS}
