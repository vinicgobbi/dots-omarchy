from post_omarchy import ui
from post_omarchy.modulos import Modulo


def criar_webapps(s):
    ui.info("Removendo os webapps de fábrica do Omarchy (omarchy-webapp-remove-all)...")
    if s.como_usuario("omarchy-webapp-remove-all", check=False).returncode != 0:
        ui.aviso("Falha ao remover os webapps do Omarchy.")

    for webapp in s.dados["webapps"]:
        # Ícone local (ex.: assets/icons/glpi.svg) é relativo à raiz do projeto.
        icone = webapp["icone"]
        if "://" not in icone:
            icone = str(s.raiz / icone)
        if s.como_usuario(["omarchy-webapp-install", webapp["nome"], webapp["url"], icone],
                          check=False, quieto=True).returncode == 0:
            ui.info(f"Webapp criado: {webapp['nome']}")
        else:
            ui.aviso(f"Não foi possível criar o webapp {webapp['nome']}.")

    ui.sucesso("Webapps configurados.")


MODULO = Modulo("webapps", "Criar webapps",
                "Apaga os webapps de fábrica do Omarchy e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, "
                "GitHub, Claude, GLPI e o admin console do Tailscale",
                criar_webapps)
