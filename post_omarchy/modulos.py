"""Registro, carga e dependências dos módulos de modules/."""

import importlib.util
from dataclasses import dataclass, field
from pathlib import Path
from typing import Callable

from . import ui


@dataclass
class Modulo:
    """Cada arquivo em modules/ define MODULO = Modulo(...) para se anunciar ao
    orquestrador. deps são ids de outros módulos que precisam rodar antes."""
    id: str
    titulo: str
    descricao: str
    executar: Callable
    deps: list[str] = field(default_factory=list)


def carregar(pasta: Path) -> list[Modulo]:
    """Carrega modules/*.py em ordem de nome (a numeração define a execução)."""
    modulos = []
    for arquivo in sorted(pasta.glob("[0-9]*.py")):
        spec = importlib.util.spec_from_file_location(f"modules.{arquivo.stem}", arquivo)
        codigo = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(codigo)
        modulos.append(codigo.MODULO)
    return modulos


def resolver_dependencias(modulos: list[Modulo], selecionados: list[bool]) -> None:
    """Marca como selecionada qualquer dependência de um módulo selecionado,
    mesmo que o usuário a tenha desmarcado, e avisa qual puxou qual. Sem isso,
    dá pra selecionar só "Extensões PHP (SQL Server)" sem "Instalar pacotes
    base" e a instalação falha no meio. Repete até estabilizar, para cobrir
    cadeias (A depende de B, que depende de C)."""
    indice = {m.id: i for i, m in enumerate(modulos)}
    mudou = True
    while mudou:
        mudou = False
        for i, modulo in enumerate(modulos):
            if not selecionados[i]:
                continue
            for dep in modulo.deps:
                j = indice.get(dep)
                if j is not None and not selecionados[j]:
                    selecionados[j] = True
                    mudou = True
                    ui.aviso(f"'{modulo.titulo}' depende de '{modulos[j].titulo}': selecionando automaticamente.")
