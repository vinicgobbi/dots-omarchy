from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_flatpaks(s):
    # O Spotify tem pacote nativo pré-compilado no repositório [omarchy]
    # (mesmo que "omarchy install service spotify" usa), então instalamos via
    # pacman em vez de duplicar via Flatpak.
    ui.info("Instalando Spotify (pacote nativo do repositório do Omarchy)...")
    s.pacman(*s.pacotes("flatpaks")["pacman"])

    ui.info("Configurando Flatpak e instalando apps...")
    s.flatpak(*s.dados["flatpaks"]["apps"])
    ui.sucesso("Aplicativos Flatpak instalados.")


MODULO = Modulo("flatpaks", "Instalar Flatpaks",
                "Spotify nativo e a lista de aplicativos Flatpak (Postman, Obsidian, etc.)",
                instalar_flatpaks, ["pacotes_base"])
