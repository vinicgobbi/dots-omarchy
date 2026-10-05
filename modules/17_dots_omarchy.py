import filecmp
import tempfile
from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo

# Arquivos que só existem para manter a pasta no git.
IGNORADOS = {".gitkeep"}


def instalar_dot(s, origem: Path, destino: str, modo: int):
    """Copia um arquivo de dots/ para ~/<destino> como o usuário-alvo,
    guardando o original em <arquivo>.bak-post-omarchy na primeira vez que ele
    difere do que vamos instalar (para nunca perder customização anterior)."""
    alvo = s.home / destino
    if alvo.is_file():
        if filecmp.cmp(origem, alvo, shallow=False):
            return
        backup = alvo.with_name(alvo.name + ".bak-post-omarchy")
        if not backup.exists():
            s.copiar_preservando(alvo, backup)

    s.instalar_arquivo(origem, alvo, modo)
    ui.info(f"Instalado ~/{destino}")


def arquivos_da_entrada(dots: Path, entrada: dict) -> list[tuple[Path, str]]:
    """(origem, destino) de uma entrada do dots.json. Com glob, o destino é uma
    pasta e cada arquivo mantém o caminho relativo à parte fixa do padrão
    ("hypr/*.lua" -> ".config/hypr/<nome>.lua")."""
    origem = entrada["origem"]
    if "*" not in origem:
        return [(dots / origem, entrada["destino"])]

    base = dots / origem.split("*")[0]
    return [(arquivo, entrada["destino"] + str(arquivo.relative_to(base)))
            for arquivo in sorted(dots.glob(origem))
            if arquivo.is_file() and arquivo.name not in IGNORADOS]


def aplicar_dots_omarchy(s):
    ui.info(f"Aplicando dots do Omarchy para {s.usuario}...")
    dots = s.raiz / "dots"

    for entrada in s.dados["dots"]:
        modo = int(entrada.get("modo", "0644"), 8)
        for origem, destino in arquivos_da_entrada(dots, entrada):
            if not origem.is_file():
                # Ex.: o CLAUDE.md global não é versionado (ver
                # dots/claude/README.md), então só é instalado se alguém
                # tiver colocado o arquivo lá.
                if entrada.get("opcional"):
                    ui.aviso(entrada.get("aviso_ausente", f"dots/{entrada['origem']} não encontrado; pulando."))
                    continue
                ui.erro(f"dots/{entrada['origem']} não encontrado.")

            if entrada.get("trocar_home"):
                # O Execute do Solaar não expande ~, então o rules.yaml traz o
                # caminho absoluto de /home/vinicius, trocado aqui pelo home do
                # usuário-alvo.
                conteudo = origem.read_text().replace("/home/vinicius/", f"{s.home}/")
                with tempfile.NamedTemporaryFile("w", suffix=origem.suffix) as temporario:
                    temporario.write(conteudo)
                    temporario.flush()
                    instalar_dot(s, Path(temporario.name), destino, modo)
            else:
                instalar_dot(s, origem, destino, modo)

    ui.sucesso("Dots do Omarchy aplicados. Recarregue o Hyprland (super+shift+r, ou faça logout/login) "
               "para ver as mudanças.")


MODULO = Modulo("dots_omarchy", "Aplicar dots do Omarchy",
                "Copia a config do Hyprland, do omarchy-shell, o screensaver, o Solaar, os scripts de "
                "~/.local/bin e o CLAUDE.md global, se houver (com backup do que existia)",
                aplicar_dots_omarchy)
