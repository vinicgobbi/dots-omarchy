from post_omarchy import ui
from post_omarchy.modulos import Modulo


def avisar_riscos_plugins(plugins):
    print(flush=True)
    print(f"{ui.NEGRITO}{ui.VERMELHO}ATENÇÃO: plugins de terceiros{ui.NC}")
    print(f"Os plugins abaixo são baixados direto do GitHub de cada autor e {ui.NEGRITO}não{ui.NC}")
    print("são revisados nem mantidos pelo Omarchy nem por este script.")
    print()
    print("  - Rodam como código arbitrário, sem sandbox, dentro do omarchy-shell,")
    print("    com as mesmas permissões do seu usuário: podem ler e alterar seus")
    print("    arquivos, executar comandos e acessar a rede.")
    print("  - É instalada a versão atual de cada repositório, que pode ter mudado")
    print("    (ou sido comprometida) desde que esta lista foi montada.")
    print("  - Um plugin com bug pode travar ou deixar lenta a barra e o shell.")
    print("  - Alguns pedem passos extras (dependências, scripts de instalação);")
    print("    veja o README de cada um.")
    print()
    print("Revise o código antes, se puder. Plugins que serão instalados:")
    for url in plugins:
        print(f"  - {url}")
    print(flush=True)


def instalar_plugins_omarchy(s):
    """Plugins são código de terceiros que roda dentro do omarchy-shell: por
    isso este é um dos últimos módulos e só instala depois de mostrar os riscos
    e o usuário confirmar. Sem terminal interativo, não instala nada."""
    plugins = s.dados["plugins"]
    if not plugins:
        ui.aviso("Nenhum plugin listado em data/plugins.json; nada a fazer.")
        return

    if not ui.terminal_interativo():
        ui.aviso("Sem terminal interativo para confirmar: plugins de terceiros NÃO foram instalados.")
        return

    avisar_riscos_plugins(plugins)
    if not ui.confirmar(f"Entendi os riscos. Instalar os {len(plugins)} plugins acima?"):
        ui.aviso("Plugins de terceiros não instalados (escolha do usuário).")
        return

    # --yes porque a confirmação já foi feita acima (e o su não tem o gum
    # interativo do omarchy). Cada plugin é best-effort: precisa de rede e
    # falha se já estiver instalado.
    for url in plugins:
        ui.info(f"Plugin do Omarchy: {url}")
        if s.como_usuario(["omarchy", "plugin", "add", url, "--yes"], check=False).returncode != 0:
            ui.aviso(f"Não foi possível instalar o plugin {url}; rode depois: omarchy plugin add {url}")

    ui.sucesso("Plugins instalados. A posição deles na barra vem do shell.json dos dots.")


MODULO = Modulo("plugins_omarchy", "Instalar plugins de terceiros",
                "Mostra os riscos e pergunta antes de instalar os plugins da barra listados em data/plugins.json",
                instalar_plugins_omarchy)
