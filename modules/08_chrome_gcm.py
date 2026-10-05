import json
import threading
import urllib.request
from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo

GCM_RELEASE_API = "https://api.github.com/repos/git-ecosystem/git-credential-manager/releases/latest"
GCM_TAR = Path("/tmp/gcm.tar.gz")
GCM_DIR = Path("/usr/local/gcm")


def url_gcm() -> str:
    """Asset .tar.gz linux-x64 (sem os símbolos) da última release."""
    with urllib.request.urlopen(GCM_RELEASE_API) as resposta:
        release = json.load(resposta)
    return next(a["browser_download_url"] for a in release["assets"]
                if a["name"].endswith(".tar.gz") and "linux-x64" in a["name"] and "symbols" not in a["name"])


def instalar_chrome_e_gcm(s):
    ui.info("Instalando Git Credential Manager e Google Chrome...")

    # GCM (tar.gz local) e Chrome são independentes: o download do GCM roda
    # em segundo plano enquanto o Chrome é instalado.
    url = url_gcm()
    download = threading.Thread(target=s.run, args=(["curl", "-sSL", "-o", str(GCM_TAR), url],))
    download.start()

    # omarchy-install-browser instala via yay e ainda configura política/tema
    # do Chrome — a política grava em /etc/opt/chrome/policies/managed, o que
    # pode pedir a senha de sudo do usuário-alvo no meio da instalação. Roda
    # em primeiro plano de propósito, pra não perder esse prompt.
    s.como_usuario("omarchy-install-browser chrome")
    download.join()

    s.run(["mkdir", "-p", str(GCM_DIR)])
    s.run(["tar", "-xzf", str(GCM_TAR), "-C", str(GCM_DIR)])
    s.link(GCM_DIR / "git-credential-manager", Path("/usr/local/bin/git-credential-manager"))
    ui.sucesso("Chrome e GCM instalados.")


MODULO = Modulo("chrome_gcm", "Chrome + Git Credential Manager",
                "Instala o Google Chrome (integrado ao tema do Omarchy) e o Git Credential Manager",
                instalar_chrome_e_gcm)
