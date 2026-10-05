import shutil
from pathlib import Path

from post_omarchy import ui
from post_omarchy.modulos import Modulo

DOTFILES_REPO = "https://github.com/vinicgobbi/Dotfiles.git"
SOLAAR_DESKTOP = (
    # Prioriza o .desktop do pacote nativo; usa o do Flatpak como alternativa.
    Path("/usr/share/applications/solaar.desktop"),
    Path("/var/lib/flatpak/exports/share/applications/io.github.pwr_solaar.solaar.desktop"),
)


def passo_usuario(s, cmd, descricao):
    """Passos de ambiente do usuário são best-effort: avisa e segue."""
    if s.como_usuario(cmd, check=False).returncode != 0:
        ui.aviso(f"Falha ao {descricao}; seguindo com o resto.")


def configurar_usuario(s):
    ui.info(f"Aplicando configurações locais para {s.usuario}...")
    s.run(["usermod", "-aG", "docker", s.usuario])

    # chsh só aceita contas presentes de fato em /etc/passwd; contas de
    # domínio (AD/LDAP via sssd/winbind) resolvem por NSS mas não estão no
    # arquivo, e o chsh recusa. A resolução restrita à fonte "files" (local)
    # distingue os dois casos.
    if s.ok(["getent", "-s", "files", "passwd", s.usuario]):
        s.run(["chsh", "-s", shutil.which("zsh") or "/usr/bin/zsh", s.usuario])
    else:
        ui.aviso(f"'{s.usuario}' é uma conta de domínio (AD/LDAP), não local: pulando 'chsh' "
                 "(não dá para trocar o shell padrão por aqui nesse caso).")

    # O tema (GTK, ícones, Nautilus etc.) já é gerido pelo próprio Omarchy
    # ("omarchy theme"), então não aplicamos gsettings/adw-gtk3/Yaru aqui.

    # Git Credential Manager
    passo_usuario(s, "git-credential-manager configure", "configurar o Git Credential Manager")
    passo_usuario(s, "git config --global credential.credentialStore secretservice",
                  "definir o credentialStore do Git")

    # Clone+bootstrap dos dotfiles (Oh My Zsh, plugins, tema, fontes e config do Solaar).
    passo_usuario(s, f"rm -rf /tmp/dotfiles && git clone {DOTFILES_REPO} /tmp/dotfiles "
                     "&& bash /tmp/dotfiles/bootstrap.sh", "aplicar os Dotfiles")

    # Node via mise, que o Omarchy já usa por padrão (o próprio Omarchy costuma
    # já ter fixado uma versão no ~/.config/mise/config.toml; só instalamos se não).
    passo_usuario(s, "mise which node &>/dev/null || mise use --global node@lts", "instalar o Node via mise")

    # Ativa o mise no zsh (o Omarchy só ativa no bash). Precisa rodar depois do
    # Oh My Zsh, que é quem cria/substitui o ~/.zshrc.
    zshrc = s.home / ".zshrc"
    if not zshrc.exists() or "mise activate zsh" not in zshrc.read_text():
        s.escrever(zshrc, 'eval "$(mise activate zsh)"\n', anexar=True, do_usuario=True)

    configurar_projetos(s)
    configurar_autostart_solaar(s)
    ui.sucesso("Ambiente de usuário configurado.")


def configurar_projetos(s):
    """Diretório de Projetos (XDG e bookmark do Nautilus), com o nome no
    idioma do usuário (o LANG vem do login shell dele)."""
    lang = s.como_usuario('printf %s "$LANG"', capturar=True, leitura=True, check=False).stdout or ""
    projetos = s.home / ("Projetos" if lang.startswith("pt_") else "Projects")
    s.criar_dir_usuario(projetos)
    passo_usuario(s, ["xdg-user-dirs-update", "--set", "PROJECTS", str(projetos)],
                  "registrar o diretório de projetos no XDG")

    bookmarks = s.home / ".config/gtk-3.0/bookmarks"
    if not bookmarks.exists() or f"file://{projetos}" not in bookmarks.read_text():
        s.escrever(bookmarks, f"file://{projetos}\n", anexar=True, do_usuario=True)


def configurar_autostart_solaar(s):
    origem = next((p for p in SOLAAR_DESKTOP if p.exists()), None)
    autostart = s.home / ".config/autostart"
    s.criar_dir_usuario(autostart)
    if origem:
        s.escrever(autostart / origem.name, origem.read_text(), do_usuario=True)


MODULO = Modulo("ambiente_usuario", "Configurar ambiente do usuário",
                "Zsh, Oh My Zsh, Node (via mise) e dotfiles de shell (o tema visual fica com o Omarchy)",
                configurar_usuario, ["pacotes_base"])
