from pathlib import PurePosixPath

from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_vscode_nautilus(s):
    ui.info("Instalando extensões do VSCode, de copiar caminho e de abrir no terminal para o Nautilus...")
    pacotes = s.pacotes("vscode_nautilus")

    s.pacman(*pacotes["pacman"])

    # Cada extensão é um repositório com install.sh próprio, rodado como o
    # usuário-alvo a partir de um clone em /tmp.
    for repo in pacotes["extensoes_git"]:
        pasta = f"/tmp/{PurePosixPath(repo).stem}"
        if s.como_usuario(f"rm -rf {pasta} && git clone {repo} {pasta} && bash {pasta}/install.sh",
                          check=False).returncode != 0:
            ui.aviso(f"Falha ao instalar a extensão do Nautilus de {repo}.")

    ui.sucesso("Extensões do VSCode, de copiar caminho e de abrir no terminal instaladas no Nautilus.")


MODULO = Modulo("vscode_nautilus", "Extensões do Nautilus (VSCode, copiar caminho e abrir no terminal)",
                "Adiciona 'Abrir com o VSCode', 'Copiar caminho' e 'Abrir no Terminal' ao menu de contexto do Nautilus",
                instalar_vscode_nautilus, ["pacotes_base"])
