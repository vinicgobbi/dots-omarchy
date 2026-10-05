from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_launchers_jogos(s):
    instalar_jogos_nativos(s)

    ui.info("Instalando launchers de jogos via Flatpak...")
    s.flatpak(*s.dados["flatpaks"]["jogos"])
    ui.sucesso("Launchers de jogos instalados.")


def instalar_jogos_nativos(s):
    """Steam (multilib) e Heroic (repositório [omarchy]) têm pacote nativo
    pronto — mesmo caminho que "omarchy install gaming steam"/"heroic" usam —
    incluindo a detecção e instalação dos drivers gráficos lib32 (Vulkan
    Intel/AMD, NVIDIA) certos pra essa máquina."""
    ui.info("Instalando Steam e Heroic (pacotes nativos, como o instalador do Omarchy faria)...")
    s.pacman(*s.pacotes("launchers_jogos")["pacman"])
    if s.comando_existe("omarchy-install-gaming-gpu-lib32"):
        s.run(["omarchy-install-gaming-gpu-lib32"])
    ui.sucesso("Steam e Heroic instalados nativamente.")


MODULO = Modulo("launchers_jogos", "Instalar launchers de jogos",
                "Steam e Heroic nativos; ProtonPlus e PrismLauncher via Flatpak",
                instalar_launchers_jogos, ["pacotes_base"])
