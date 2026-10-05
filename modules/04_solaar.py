from post_omarchy import ui
from post_omarchy.modulos import Modulo

REGRAS_UDEV_URL = ("https://raw.githubusercontent.com/pwr-Solaar/Solaar/master/"
                   "rules.d-uinput/42-logitech-unify-permissions.rules")


def configurar_solaar(s):
    ui.info("Instalando Solaar...")
    pacotes = s.pacotes("solaar")

    if s.ok(["pacman", "-Si", *pacotes["pacman"]]):
        s.pacman(*pacotes["pacman"])
        ui.sucesso("Solaar instalado via pacote nativo (pacman).")
    else:
        ui.aviso("Pacote nativo do Solaar não encontrado; instalando via Flatpak.")
        s.flatpak(pacotes["flatpak_alternativo"])
        ui.sucesso("Solaar instalado via Flatpak.")

    if not s.grupo_existe("plugdev"):
        s.run(["groupadd", "plugdev"])
    s.run(["usermod", "-aG", "plugdev", s.usuario])

    ui.info("Configurando regras UDEV para o Solaar...")
    s.run(["curl", "-sL", REGRAS_UDEV_URL, "-o", "/etc/udev/rules.d/42-logitech-unify-permissions.rules"])

    s.run(["udevadm", "control", "--reload-rules"])
    s.run(["udevadm", "trigger", "--subsystem-match=usb"])
    s.run(["udevadm", "trigger", "--subsystem-match=hidraw"])

    ui.sucesso("Regras UDEV do Solaar configuradas.")


MODULO = Modulo("solaar", "Instalar e configurar Solaar",
                "Instala o Solaar (nativo, com fallback para Flatpak) e as regras UDEV do receptor Logitech Unifying",
                configurar_solaar, ["pacotes_base"])
