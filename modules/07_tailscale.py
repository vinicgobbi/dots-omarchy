from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_tailscale(s):
    ui.info("Instalando o Tailscale...")

    # tailscale é pacote oficial do repo "extra" — mesmo pacote que
    # "omarchy install service tailscale" usa.
    s.pacman(*s.pacotes("tailscale")["pacman"])

    s.run(["systemctl", "enable", "--now", "tailscaled"])
    s.run(["tailscale", "set", f"--operator={s.usuario}"])

    integrar_tailscale_omarchy(s)

    ui.sucesso(f"Tailscale instalado e '{s.usuario}' definido como operator.")


def integrar_tailscale_omarchy(s):
    """Replica o que "omarchy-install-service-tailscale" faz: widget na barra,
    atalho de webapp pro admin console e recebimento de Taildrop em
    ~/Downloads. Cada etapa é best-effort (não aborta o módulo): a sessão
    gráfica do usuário-alvo pode não estar ativa ainda nesse momento."""
    runtime_dir = f"/run/user/{s.uid}"

    if s.como_usuario(f"XDG_RUNTIME_DIR='{runtime_dir}' systemctl --user enable --now "
                      "omarchy-tailscale-receive.service", check=False, quieto=True).returncode == 0:
        ui.info("Recebimento de Taildrop habilitado em ~/Downloads.")
    else:
        ui.aviso(f"Não foi possível habilitar o recebimento de Taildrop agora (sessão gráfica de "
                 f"'{s.usuario}' inativa?); rode depois de logar: "
                 "systemctl --user enable --now omarchy-tailscale-receive.service")

    if s.como_usuario("omarchy-plugin-enable omarchy.tailscale", check=False, quieto=True).returncode == 0:
        ui.info("Tailscale adicionado à barra do Omarchy.")
    else:
        ui.aviso("Não foi possível adicionar o Tailscale à barra do Omarchy agora.")

    criar_webapp_tailscale_admin(s)


def criar_webapp_tailscale_admin(s):
    """O módulo "webapps" também cria este atalho (a entrada vem do mesmo
    data/webapps.json), porque o omarchy-webapp-remove-all apaga todos os
    webapps de uma vez."""
    webapp = next(w for w in s.dados["webapps"] if w["nome"] == "Tailscale")
    if s.como_usuario(["omarchy-webapp-install", webapp["nome"], webapp["url"], webapp["icone"]],
                      check=False, quieto=True).returncode != 0:
        ui.aviso("Não foi possível criar o atalho de webapp do Tailscale Admin Console.")


MODULO = Modulo("tailscale", "Instalar Tailscale",
                "Tailscale (VPN mesh) com widget na barra e webapp do admin console",
                instalar_tailscale)
