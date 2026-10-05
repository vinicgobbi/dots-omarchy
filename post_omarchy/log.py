"""Log em arquivo.

Cada execução grava dois arquivos em logs/, ao lado do script:
  <prefixo>_<timestamp>.log      — só as mensagens info/aviso/erro/sucesso/passo,
                                   com hora e nível, sem código de cor.
  <prefixo>_<timestamp>.raw.log  — transcrição bruta e completa de tudo que
                                   passa pelo terminal a partir de iniciar_bruto
                                   (inclui a saída de pacman/yay/curl/flatpak etc.).
"""

import atexit
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path

_arquivo: Path | None = None
_arquivo_bruto: Path | None = None


def iniciar(raiz: Path, prefixo: str) -> Path:
    """Prepara só o log "limpo"; o bruto começa depois, em iniciar_bruto."""
    global _arquivo, _arquivo_bruto
    pasta = raiz / "logs"
    pasta.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y-%m-%d_%H%M%S")
    _arquivo = pasta / f"{prefixo}_{timestamp}.log"
    _arquivo_bruto = pasta / f"{prefixo}_{timestamp}.raw.log"
    _arquivo.write_text("")
    return _arquivo


def linha(nivel: str, mensagem: str) -> None:
    """Grava uma linha com timestamp e nível. Chamada pelas mensagens de ui."""
    if _arquivo is None:
        return
    with _arquivo.open("a") as f:
        f.write(f"{datetime.now():%Y-%m-%d %H:%M:%S} [{nivel}] {mensagem}\n")


def iniciar_bruto() -> Path | None:
    """A partir daqui, TUDO que passar pelo terminal (stdout e stderr, inclusive
    a saída dos comandos filhos) também vai para o log bruto.

    Mesma técnica do `exec > >(tee -a ...)` do bash: os fds 1 e 2 do processo
    passam a apontar para a entrada de um `tee`, que herdou o terminal original.
    Só é chamada depois do menu interativo, de propósito: com o stdout
    redirecionado, isatty() passa a dizer "não é terminal", o que quebraria a
    detecção de terminal do menu se isso rodasse antes dele.
    """
    if _arquivo_bruto is None:
        return None
    sys.stdout.flush()
    sys.stderr.flush()
    tee = subprocess.Popen(["tee", "-a", str(_arquivo_bruto)], stdin=subprocess.PIPE)
    stdout_original, stderr_original = os.dup(1), os.dup(2)
    os.dup2(tee.stdin.fileno(), 1)
    os.dup2(tee.stdin.fileno(), 2)

    def encerrar() -> None:
        # Devolve o terminal e fecha a entrada do tee para ele terminar de
        # gravar antes de o script sair.
        sys.stdout.flush()
        sys.stderr.flush()
        os.dup2(stdout_original, 1)
        os.dup2(stderr_original, 2)
        tee.stdin.close()
        tee.wait()

    atexit.register(encerrar)
    return _arquivo_bruto
