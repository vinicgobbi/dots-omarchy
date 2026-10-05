from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_pacotes_base(s):
    ui.info("Instalando pacotes base e utilitários...")
    # Docker, PHP/composer e unixODBC são oficiais (repo extra/core). O
    # repositório [omarchy] já traz o visual-studio-code-bin pré-compilado
    # (mesmo pacote que "omarchy install editor vscode" usa); msodbcsql e
    # mssql-tools só existem na AUR (resolvidos pelo Chaotic-AUR, sem compilar).
    pacotes = s.pacotes("pacotes_base")
    s.pacman(*pacotes["pacman"])
    s.aur(*pacotes["aur"])
    aplicar_config_omarchy_vscode(s)

    # O pacote da AUR instala em /opt/mssql-tools (sem versão).
    s.escrever(Path("/etc/profile.d/mssql-tools.sh"), 'export PATH="$PATH:/opt/mssql-tools/bin"\n')
    ui.sucesso("Pacotes e MS SQL instalados.")


def aplicar_config_omarchy_vscode(s):
    """Replica o que "omarchy-install-editor-vscode" faz para o usuário-alvo
    (senha via gnome-libsecret, autoupdate desligado — o Omarchy já atualiza o
    VSCode via pacman — e o tema atual do Omarchy aplicado no editor), sem o
    lançamento automático da GUI ao final, que assume uma sessão gráfica
    interativa e não faz sentido no meio de uma instalação em lote."""
    s.escrever(s.home / ".vscode/argv.json", '{\n  "password-store":"gnome-libsecret"\n}\n', do_usuario=True)
    s.escrever(s.home / ".config/Code/User/settings.json", '{\n  "update.mode": "none"\n}\n', do_usuario=True)
    s.como_usuario("omarchy-theme-set-vscode")


MODULO = Modulo("pacotes_base", "Instalar pacotes base",
                "Docker, VSCode, PHP, Zsh e ferramentas de SQL Server",
                instalar_pacotes_base, ["repositorios"])
