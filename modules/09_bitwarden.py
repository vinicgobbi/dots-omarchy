from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_bitwarden_nativo(s):
    ui.info("Instalando Bitwarden nativo (integração com navegador)...")
    s.aur(*s.pacotes("bitwarden")["aur"])
    ui.sucesso("Bitwarden instalado nativamente.")


MODULO = Modulo("bitwarden", "Instalar Bitwarden nativo",
                "Instala o Bitwarden desktop (necessário para integração nativa com o navegador)",
                instalar_bitwarden_nativo, ["repositorios"])
