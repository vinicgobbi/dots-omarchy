import re

from post_omarchy import ui
from post_omarchy.modulos import Modulo

DNS = re.compile(r"^\s*dhcp-option\s+DNS\s+(\S+)", re.MULTILINE)
DOMINIO = re.compile(r"^\s*dhcp-option\s+DOMAIN(?:-SEARCH)?\s+(\S+)", re.MULTILINE)


def importar_ovpn(s):
    ui.info("Importando perfis OpenVPN para o NetworkManager...")

    plugin = s.pacotes("ovpn")["pacman"]
    if not s.ok(["pacman", "-Qq", *plugin]):
        s.pacman(*plugin)

    pasta = s.raiz / "OVPN"
    arquivos = sorted(pasta.glob("*.ovpn"))
    if not arquivos:
        ui.aviso(f"Nenhum arquivo .ovpn encontrado em {pasta} — pulando importação.")
        return

    existentes = s.run(["nmcli", "-g", "NAME", "connection", "show"],
                       capturar=True, leitura=True).stdout.splitlines()
    for arquivo in arquivos:
        nome = arquivo.stem

        # Reimportar do zero para o script continuar idempotente em reexecuções.
        if nome in existentes:
            s.run(["nmcli", "connection", "delete", nome], quieto=True)

        if s.run(["nmcli", "connection", "import", "type", "openvpn", "file", str(arquivo)],
                 check=False, quieto=True).returncode != 0:
            ui.aviso(f"Falha ao importar {arquivo}, pulando.")
            continue

        # DNS e domínio de busca não vêm sempre preenchidos pela importação
        # automática do plugin; quando o .ovpn declara essas diretivas, aplica
        # manualmente via nmcli.
        conteudo = arquivo.read_text()
        dns = " ".join(DNS.findall(conteudo))
        dominios = " ".join(DOMINIO.findall(conteudo))
        if dns:
            s.run(["nmcli", "connection", "modify", nome, "ipv4.dns", dns])
        if dominios:
            s.run(["nmcli", "connection", "modify", nome, "ipv4.dns-search", dominios])

        ui.sucesso(f"Perfil '{nome}' importado.")


MODULO = Modulo("ovpn", "Importar perfis OpenVPN",
                "Instala o plugin OpenVPN do NetworkManager e importa os .ovpn de ./OVPN "
                "(aplicando DNS/domínio quando presentes no arquivo)",
                importar_ovpn)
