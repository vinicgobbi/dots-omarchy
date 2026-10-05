#!/usr/bin/env python3
"""Pós-instalação do Omarchy: menu de módulos, dependências, execução e log."""

import argparse
import os
import sys
import time
from pathlib import Path

from post_omarchy import dados, log, modulos, ui
from post_omarchy.sistema import FalhaComando, Sistema

RAIZ = Path(__file__).resolve().parent


def detectar_sistema() -> None:
    """Este projeto é exclusivo do Omarchy (Arch por baixo: pacman, yay
    pré-instalado, Hyprland)."""
    release = {}
    for linha in Path("/etc/os-release").read_text().splitlines():
        chave, _, valor = linha.partition("=")
        release[chave] = valor.strip('"')
    nome = release.get("PRETTY_NAME", "desconhecido")
    if release.get("ID") != "omarchy":
        ui.erro(f"Sistema não suportado: {nome}. Este script é exclusivo do Omarchy (ID=omarchy). "
                "Para outras distros, use o projeto post_install.")
    ui.info(f"Sistema detectado: {nome} (Omarchy — Arch por baixo, pacman + Chaotic-AUR/yay)")


def argumentos() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--dry-run", action="store_true",
                        help="mostra os comandos e arquivos que seriam alterados, sem executar")
    parser.add_argument("--modulos", metavar="ID,ID",
                        help="roda só estes módulos (e as dependências), sem o menu")
    return parser.parse_args()


def main() -> None:
    args = argumentos()
    inicio = time.monotonic()
    ui.banner()

    if os.geteuid() != 0 and not args.dry_run:
        ui.erro("Execute este script como root (sudo).")

    ui.info(f"Log desta execução: {log.iniciar(RAIZ, 'setup')}")
    detectar_sistema()

    lista = modulos.carregar(RAIZ / "modules")
    sistema = Sistema(RAIZ, ui.perguntar_usuario(), dados.carregar(RAIZ), args.dry_run)
    ui.info(f"Configuração será aplicada para: {sistema.usuario}")
    if args.dry_run:
        ui.aviso("Modo --dry-run: nada será alterado no sistema.")

    if args.modulos:
        pedidos = set(args.modulos.split(","))
        desconhecidos = pedidos - {m.id for m in lista}
        if desconhecidos:
            ui.erro(f"Módulos desconhecidos: {', '.join(sorted(desconhecidos))}")
        selecionados = [m.id in pedidos for m in lista]
    else:
        selecionados = ui.menu_checklist(
            "Escolha o que deseja executar (o script já vem com tudo marcado):",
            [(m.titulo, m.descricao) for m in lista])

    # Alguns módulos dependem de outro ter rodado antes (ex.: "Extensões PHP"
    # precisa dos headers do driver ODBC que só "Instalar pacotes base" traz).
    modulos.resolver_dependencias(lista, selecionados)
    escolhidos = [m for m, sel in zip(lista, selecionados) if sel]

    if not escolhidos:
        ui.aviso("Nenhum módulo selecionado. Nada a fazer.")
        return

    print(flush=True)
    ui.info("Módulos selecionados:")
    for modulo in escolhidos:
        print(f"  - {modulo.titulo}")
    print(flush=True)

    if ui.terminal_interativo() and not ui.confirmar(
            f"Iniciar a configuração com os {len(escolhidos)} módulos acima?"):
        ui.aviso("Operação cancelada pelo usuário.")
        return

    bruto = log.iniciar_bruto()
    if bruto:
        ui.info(f"Saída completa (bruta) desta execução: {bruto}")

    executados = []
    for i, modulo in enumerate(escolhidos, start=1):
        ui.passo(i, len(escolhidos), modulo.titulo)
        modulo.executar(sistema)
        executados.append(modulo.titulo)

    ui.resumo_final(executados, int(time.monotonic() - inicio))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print(f"\n{ui.VERMELHO}[-] Cancelado pelo usuário.{ui.NC}", flush=True)
        sys.exit(130)
    except FalhaComando as falha:
        # Equivale ao "set -e" + trap ERR do bash: qualquer comando que falhe
        # fora dos passos best-effort interrompe o setup.
        ui.erro(f"Falha inesperada: {falha.comando}", falha.codigo)
