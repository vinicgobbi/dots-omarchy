from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo

SUDOERS_AUR_FILE = Path("/etc/sudoers.d/99-post-omarchy-aur")


def limpeza_final(s):
    ui.info("Limpando o sistema...")
    orfaos = s.run(["pacman", "-Qtdq"], check=False, capturar=True, leitura=True).stdout.split()
    if orfaos:
        s.run(["pacman", "-Rns", "--noconfirm", *orfaos])
    s.run(["pacman", "-Sc", "--noconfirm"])

    if SUDOERS_AUR_FILE.exists():
        s.remover(SUDOERS_AUR_FILE)
        ui.info("Regra temporária de sudo sem senha para o pacman (liberada para o yay) removida.")

    # Mesmo alcance do "rm -rf /tmp/*": não mexe em entradas ocultas.
    for item in Path("/tmp").glob("*"):
        try:
            s.remover(item)
        except OSError:
            pass

    ui.sucesso("Instalação finalizada com sucesso!")
    ui.info("Recomenda-se reiniciar a máquina para aplicar as mudanças de grupo e kernel.")


MODULO = Modulo("limpeza", "Limpeza final",
                "Remove pacotes órfãos, limpa o cache do pacman e remove o sudo temporário do yay",
                limpeza_final)
