"""UI: cores, mensagens, prompts e menu interativo."""

import os
import select
import sys
import termios
import time

from . import log

VERDE = "\033[0;32m"
AMARELO = "\033[1;33m"
VERMELHO = "\033[0;31m"
AZUL = "\033[0;34m"
CINZA = "\033[1;30m"
NEGRITO = "\033[1m"
NC = "\033[0m"


def _escrever(texto: str) -> None:
    # flush sempre: a saída dos comandos filhos vai direto para o fd 1, então
    # o buffer do Python precisa estar vazio para manter a ordem na tela/log.
    print(texto, flush=True)


def info(msg: str) -> None:
    _escrever(f"{AMARELO}[*] {msg}{NC}")
    log.linha("INFO", msg)


def sucesso(msg: str) -> None:
    _escrever(f"{VERDE}[+] {msg}{NC}")
    log.linha("OK", msg)


def aviso(msg: str) -> None:
    _escrever(f"{AZUL}[!] {msg}{NC}")
    log.linha("AVISO", msg)


def erro(msg: str, codigo: int = 1):
    """Mostra o erro, grava no log e encerra o script."""
    _escrever(f"{VERMELHO}[-] {msg}{NC}")
    log.linha("ERRO", msg)
    sys.exit(codigo)


def passo(atual: int, total: int, titulo: str) -> None:
    """Imprime "[atual/total] título" para marcar o progresso dos módulos."""
    _escrever(f"{NEGRITO}{AZUL}[{atual}/{total}]{NC} {titulo}")
    log.linha("PASSO", f"[{atual}/{total}] {titulo}")


def banner() -> None:
    _escrever(f"{NEGRITO}{VERDE}")
    _escrever("==============================================")
    _escrever("   POST-INSTALL SETUP")
    _escrever("==============================================")
    _escrever(NC)


def terminal_interativo() -> bool:
    return sys.stdin.isatty()


def confirmar(pergunta: str) -> bool:
    """Pergunta s/N genérica."""
    try:
        resposta = input(f"{AMARELO}[?] {pergunta} [s/N]: {NC}")
    except EOFError:
        return False
    return resposta.strip() in ("s", "S", "y", "Y")


def perguntar_usuario() -> str:
    """Pergunta o usuário alvo do sistema, sem valor padrão, validando que existe."""
    import pwd

    while True:
        nome = input(f"{AMARELO}[?] Para qual usuário do sistema devo configurar o ambiente? {NC}").strip()
        if not nome:
            aviso("Digite um nome de usuário.")
            continue
        try:
            pwd.getpwnam(nome)
            return nome
        except KeyError:
            aviso(f"Usuário '{nome}' não existe neste sistema. Tente novamente.")


def _limpar_tela() -> None:
    sys.stdout.write("\033[H\033[2J\033[3J")
    sys.stdout.flush()


def _ler_tecla() -> str:
    """Lê uma tecla sem eco e sem esperar enter. Setas viram "UP"/"DOWN"."""
    fd = sys.stdin.fileno()
    antigo = termios.tcgetattr(fd)
    novo = termios.tcgetattr(fd)
    novo[3] &= ~(termios.ICANON | termios.ECHO)
    try:
        termios.tcsetattr(fd, termios.TCSADRAIN, novo)
        tecla = os.read(fd, 1).decode(errors="ignore")
        if tecla == "\x1b":
            resto = b""
            fim = time.monotonic() + 0.01
            while len(resto) < 2 and select.select([fd], [], [], max(0, fim - time.monotonic()))[0]:
                resto += os.read(fd, 1)
            return {b"[A": "UP", b"[B": "DOWN"}.get(resto, "")
        return tecla
    finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, antigo)


def menu_checklist(titulo: str, itens: list[tuple[str, str]]) -> list[bool]:
    """Menu de checklist (titulo, descricao). Retorna a seleção, tudo marcado
    por padrão. Sem terminal interativo, seleciona tudo sem perguntar."""
    selecionados = [True] * len(itens)
    if not (sys.stdin.isatty() and sys.stdout.isatty()):
        aviso("Entrada não é um terminal interativo: executando todos os módulos automaticamente.")
        return selecionados

    cursor = 0
    while True:
        _limpar_tela()
        linhas = [f"{NEGRITO}{titulo}{NC}",
                  f"{CINZA}↑/↓ mover   espaço alterna   a marca/desmarca tudo   enter confirma{NC}", ""]
        for i, (nome, descricao) in enumerate(itens):
            linha = f"[{'x' if selecionados[i] else ' '}] {nome}"
            if i == cursor:
                linhas.append(f"{VERDE}> {linha}{NC} {CINZA}- {descricao}{NC}")
            else:
                linhas.append(f"  {linha} {CINZA}- {descricao}{NC}")
        print("\n".join(linhas), flush=True)

        tecla = _ler_tecla()
        if tecla == "UP" and cursor > 0:
            cursor -= 1
        elif tecla == "DOWN" and cursor < len(itens) - 1:
            cursor += 1
        elif tecla == " ":
            selecionados[cursor] = not selecionados[cursor]
        elif tecla in ("a", "A"):
            selecionados = [not all(selecionados)] * len(itens)
        elif tecla in ("\n", "\r", ""):
            _limpar_tela()
            return selecionados


def resumo_final(executados: list[str], segundos: int) -> None:
    _escrever("")
    _escrever(f"{NEGRITO}{VERDE}================ RESUMO ================{NC}")
    for item in executados:
        _escrever(f"  {VERDE}✓{NC} {item}")
    _escrever(f"{NEGRITO}Tempo total: {segundos // 60}m{segundos % 60}s{NC}")
    _escrever(f"{NEGRITO}{VERDE}========================================={NC}")
