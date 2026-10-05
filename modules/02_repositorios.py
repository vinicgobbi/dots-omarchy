from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo

# Chave GPG oficial do Chaotic-AUR (aur.chaotic.cx) — repo binário que serve
# pacotes pré-compilados do AUR, evitando compilar tudo na mão.
CHAOTIC_AUR_KEY = "3056513887B78AEB"
SUDOERS_AUR_FILE = Path("/etc/sudoers.d/99-post-omarchy-aur")
PACMAN_CONF = Path("/etc/pacman.conf")


def configurar_chaotic_aur(s):
    """Habilita o Chaotic-AUR seguindo o processo oficial
    (https://aur.chaotic.cx/): importa a chave, instala keyring+mirrorlist via
    pacman -U e inclui o mirrorlist no pacman.conf. Idempotente."""
    if any(linha.startswith("[chaotic-aur]") for linha in PACMAN_CONF.read_text().splitlines()):
        ui.info("Chaotic-AUR já configurado.")
        return

    ui.info("Configurando o repositório binário Chaotic-AUR...")
    s.run(["pacman-key", "--recv-key", CHAOTIC_AUR_KEY, "--keyserver", "keyserver.ubuntu.com"])
    s.run(["pacman-key", "--lsign-key", CHAOTIC_AUR_KEY])
    s.run(["pacman", "-U", "--noconfirm",
           "https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-keyring.pkg.tar.zst",
           "https://cdn-mirror.chaotic.cx/chaotic-aur/chaotic-mirrorlist.pkg.tar.zst"])
    s.escrever(PACMAN_CONF, "\n[chaotic-aur]\nInclude = /etc/pacman.d/chaotic-mirrorlist\n", anexar=True)
    s.run(["pacman", "-Sy"])
    ui.sucesso("Chaotic-AUR configurado.")


def permitir_sudo_pacman_temporario(s):
    """O yay (e o makepkg por baixo) recusa rodar como root, mas o setup roda
    via sudo. Sem essa liberação, todo módulo que instala algo da AUR pararia
    pedindo senha no meio. Regra restrita (só o pacman, não ALL) e temporária:
    removida em "Limpeza final" (modules/19_limpeza.py)."""
    if SUDOERS_AUR_FILE.exists():
        return

    ui.info(f"Liberando sudo sem senha para o pacman (usuário {s.usuario}) — necessário para o yay "
            "instalar pacotes da AUR sem interação; revertido ao final em 'Limpeza final'.")
    s.escrever(SUDOERS_AUR_FILE, f"{s.usuario} ALL=(ALL) NOPASSWD: /usr/bin/pacman\n", modo=0o440)
    if s.run(["visudo", "-cf", str(SUDOERS_AUR_FILE)], check=False).returncode != 0:
        s.remover(SUDOERS_AUR_FILE)
        ui.erro("Falha ao validar a regra temporária de sudo para o pacman.")


def configurar_repositorios(s):
    ui.info("Configurando repositórios (Chaotic-AUR + yay)...")

    # O mirrorlist do Omarchy já aponta para o CDN próprio dele
    # (stable-mirror.omarchy.org), que também serve o repositório [omarchy];
    # por isso não rodamos o reflector aqui. O yay já vem pré-instalado.
    s.run(["pacman", "-Sy", "--needed", "--noconfirm", *s.pacotes("repositorios")["pacman"]])

    configurar_chaotic_aur(s)
    permitir_sudo_pacman_temporario(s)

    if not s.comando_existe("yay"):
        ui.erro("yay não encontrado — deveria vir pré-instalado no Omarchy.")
    ui.sucesso("Repositórios configurados.")


MODULO = Modulo("repositorios", "Configurar repositórios",
                "Habilita o Chaotic-AUR e libera o yay para instalar pacotes da AUR",
                configurar_repositorios)
