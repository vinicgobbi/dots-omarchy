from post_omarchy import ui
from post_omarchy.modulos import Modulo


def instalar_rust_tools(s):
    ui.info("Instalando rustup e compilando eza/topgrade via cargo...")
    pacotes = s.pacotes("rust_tools")

    s.pacman(*pacotes["pacman"])

    s.como_usuario("curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs "
                   "| sh -s -- -y --default-toolchain stable")
    s.como_usuario('"$HOME/.cargo/bin/cargo" install --locked ' + " ".join(pacotes["cargo"]))
    ui.sucesso(f"rustup, eza e topgrade instalados para {s.usuario}.")


MODULO = Modulo("rust_tools", "Instalar Rust, eza e topgrade",
                "rustup e compilação de eza e topgrade via cargo",
                instalar_rust_tools)
