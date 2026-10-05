from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_temas_omarchy(s):
    """Temas extras vêm do GitHub de cada autor e podem trazer hooks
    (theme-set.d) que rodam como o usuário a cada troca de tema; por isso, como
    os plugins, só instala depois de confirmar. Sem terminal interativo, não
    instala nada."""
    temas = s.dados["themes"]
    if not temas:
        ui.aviso("Nenhum tema listado em data/themes.json; nada a fazer.")
        return

    if not ui.terminal_interativo():
        ui.aviso("Sem terminal interativo para confirmar: temas de terceiros NÃO foram instalados.")
        return

    print(flush=True)
    print("Temas que serão instalados (código de terceiros, podem trazer hooks")
    print("que rodam a cada troca de tema):")
    for url in temas:
        print(f"  - {url}")
    print(flush=True)
    if not ui.confirmar(f"Instalar os {len(temas)} temas acima?"):
        ui.aviso("Temas não instalados (escolha do usuário).")
        return

    # O omarchy-theme-install clona e já aplica o tema, então o último da
    # lista é o que fica ativo. Fora da sessão gráfica a aplicação pode
    # falhar, mas o clone fica feito e o tema aparece no seletor.
    for url in temas:
        ui.info(f"Tema do Omarchy: {url}")
        if s.como_usuario(["omarchy-theme-install", url], check=False).returncode != 0:
            ui.aviso(f"Não foi possível instalar/aplicar o tema {url}; rode depois: omarchy-theme-install {url}")

    ui.sucesso("Temas instalados. Troque pelo seletor de temas do Omarchy, se quiser.")


MODULO = Modulo("temas_omarchy", "Instalar temas extras",
                "Pergunta antes de instalar os temas listados em data/themes.json (o último fica ativo)",
                instalar_temas_omarchy)
