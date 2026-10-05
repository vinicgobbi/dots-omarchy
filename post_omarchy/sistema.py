"""Execução de comandos e escrita de arquivos, como root ou como o usuário-alvo.

Tudo que altera o sistema passa por aqui, para o --dry-run conseguir só
mostrar o que seria feito.
"""

import grp
import os
import pwd
import shlex
import shutil
import subprocess
from pathlib import Path

from . import ui


class FalhaComando(Exception):
    """Comando terminou com código diferente de zero (equivale ao set -e)."""

    def __init__(self, comando: str, codigo: int):
        super().__init__(f"{comando} (código {codigo})")
        self.comando = comando
        self.codigo = codigo


Comando = list[str] | str


def _texto(cmd: Comando) -> str:
    return cmd if isinstance(cmd, str) else shlex.join(cmd)


class Sistema:
    def __init__(self, raiz: Path, usuario: str, dados: dict, dry_run: bool = False):
        self.raiz = raiz
        self.usuario = usuario
        self.dados = dados
        self.dry_run = dry_run
        entrada = pwd.getpwnam(usuario)
        self.uid = entrada.pw_uid
        self.gid = entrada.pw_gid
        self.home = Path(entrada.pw_dir)

    # ------------------------------------------------------------------
    # Comandos
    # ------------------------------------------------------------------
    def run(self, cmd: Comando, *, check: bool = True, quieto: bool = False,
            capturar: bool = False, leitura: bool = False, env: dict | None = None,
            ) -> subprocess.CompletedProcess:
        """Roda um comando como root. String roda via shell; lista, direto.

        check:    levanta FalhaComando se o código for diferente de zero.
        quieto:   descarta stdout/stderr (equivale ao &>/dev/null).
        capturar: devolve o stdout em .stdout (texto).
        leitura:  comando só de consulta, roda mesmo no --dry-run.
        """
        texto = _texto(cmd)
        if self.dry_run and not leitura:
            print(f"{ui.CINZA}[dry-run]{ui.NC} {texto}", flush=True)
            return subprocess.CompletedProcess(cmd, 0, stdout="")

        saida = subprocess.PIPE if capturar else (subprocess.DEVNULL if quieto else None)
        resultado = subprocess.run(
            cmd, shell=isinstance(cmd, str), text=True, stdout=saida,
            stderr=subprocess.DEVNULL if quieto else None,
            env={**os.environ, **env} if env else None,
        )
        if check and resultado.returncode != 0:
            raise FalhaComando(texto, resultado.returncode)
        return resultado

    def como_usuario(self, cmd: Comando, **kwargs) -> subprocess.CompletedProcess:
        """Roda como o usuário-alvo em login shell (su -), com $HOME correto.
        Se já somos esse usuário (--dry-run sem sudo), usa um login shell direto."""
        if os.geteuid() == self.uid:
            return self.run(["bash", "-lc", _texto(cmd)], **kwargs)
        return self.run(["su", "-", self.usuario, "-c", _texto(cmd)], **kwargs)

    def ok(self, cmd: Comando, *, como_usuario: bool = False) -> bool:
        """Consulta silenciosa: True se o comando terminar com sucesso."""
        rodar = self.como_usuario if como_usuario else self.run
        return rodar(cmd, check=False, quieto=True, leitura=True).returncode == 0

    def pacman(self, *pacotes: str) -> None:
        if pacotes:
            self.run(["pacman", "-S", "--needed", "--noconfirm", *pacotes])

    def aur(self, *pacotes: str) -> None:
        """Instala via yay como o usuário-alvo (o makepkg recusa rodar como
        root). Depende do módulo "repositorios" ter liberado sudo sem senha
        para o pacman nesse usuário antes."""
        if pacotes:
            self.como_usuario(["yay", "-S", "--needed", "--noconfirm", *pacotes])

    def flatpak(self, *apps: str) -> None:
        self.run(["flatpak", "remote-add", "--if-not-exists", "flathub",
                  "https://flathub.org/repo/flathub.flatpakrepo"])
        if apps:
            self.run(["flatpak", "install", "-y", "flathub", *apps])

    @staticmethod
    def comando_existe(nome: str) -> bool:
        return shutil.which(nome) is not None

    @staticmethod
    def grupo_existe(nome: str) -> bool:
        try:
            grp.getgrnam(nome)
            return True
        except KeyError:
            return False

    def pacotes(self, modulo: str) -> dict:
        """Entrada do módulo em data/packages.json."""
        return self.dados["packages"].get(modulo, {})

    # ------------------------------------------------------------------
    # Arquivos
    # ------------------------------------------------------------------
    def _mostrar(self, acao: str, caminho: Path) -> bool:
        if self.dry_run:
            print(f"{ui.CINZA}[dry-run]{ui.NC} {acao} {caminho}", flush=True)
        return self.dry_run

    def criar_dir_usuario(self, caminho: Path) -> None:
        """mkdir -p como o usuário: cada diretório criado fica com o dono certo."""
        if self._mostrar("mkdir -p (usuário)", caminho):
            return
        faltando = []
        atual = caminho
        while not atual.exists():
            faltando.append(atual)
            atual = atual.parent
        for pasta in reversed(faltando):
            pasta.mkdir()
            os.chown(pasta, self.uid, self.gid)

    def escrever(self, caminho: Path, conteudo: str, *, modo: int = 0o644,
                 do_usuario: bool = False, anexar: bool = False) -> None:
        """Escreve (ou anexa) um arquivo. do_usuario: dono passa a ser o
        usuário-alvo, como se ele mesmo tivesse criado."""
        if self._mostrar("anexar em" if anexar else "escrever", caminho):
            return
        if do_usuario:
            self.criar_dir_usuario(caminho.parent)
        novo = not caminho.exists()
        with caminho.open("a" if anexar else "w") as f:
            f.write(conteudo)
        if novo:
            os.chmod(caminho, modo)
        if do_usuario:
            os.chown(caminho, self.uid, self.gid)

    def instalar_arquivo(self, origem: Path, destino: Path, modo: int) -> None:
        """Equivale ao `install -o USER -g GRUPO -m MODO origem destino`."""
        if self._mostrar(f"install -m {modo:o} {origem} ->", destino):
            return
        self.criar_dir_usuario(destino.parent)
        shutil.copyfile(origem, destino)
        os.chown(destino, self.uid, self.gid)
        os.chmod(destino, modo)

    def copiar_preservando(self, origem: Path, destino: Path) -> None:
        """Equivale ao `cp -p`."""
        if self._mostrar(f"cp -p {origem} ->", destino):
            return
        shutil.copy2(origem, destino)
        st = origem.stat()
        os.chown(destino, st.st_uid, st.st_gid)

    def remover(self, caminho: Path) -> None:
        """rm -rf."""
        if self._mostrar("rm -rf", caminho):
            return
        if caminho.is_dir() and not caminho.is_symlink():
            shutil.rmtree(caminho, ignore_errors=True)
        else:
            caminho.unlink(missing_ok=True)

    def link(self, alvo: Path, link: Path) -> None:
        """ln -sf."""
        if self._mostrar(f"ln -sf {alvo} ->", link):
            return
        link.unlink(missing_ok=True)
        link.symlink_to(alvo)
