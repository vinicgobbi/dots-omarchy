from post_omarchy import ui
from post_omarchy.modulos import Modulo


def configurar_extensoes_php(s):
    ui.info("Configurando extensões PHP (sqlsrv)...")

    # O php oficial do Arch não traz pecl/pear; php-sqlsrv/php-pdo_sqlsrv da
    # AUR já vêm pré-compilados (via Chaotic-AUR) e registram o próprio .ini
    # em /etc/php/conf.d, então não precisa habilitar nada à mão aqui.
    s.aur(*s.pacotes("php_extensoes")["aur"])
    ui.sucesso("Extensões PHP configuradas.")


MODULO = Modulo("php_extensoes", "Extensões PHP (SQL Server)",
                "Instala sqlsrv/pdo_sqlsrv (pré-compilados via Chaotic-AUR)",
                configurar_extensoes_php, ["pacotes_base"])
