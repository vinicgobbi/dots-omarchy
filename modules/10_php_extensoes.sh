#!/usr/bin/env bash

configurar_extensoes_php() {
    info "Configurando extensões PHP (sqlsrv)..."

    # O php oficial do Arch não traz pecl/pear; php-sqlsrv/php-pdo_sqlsrv da
    # AUR já vêm pré-compilados (via Chaotic-AUR) e registram o próprio .ini
    # em /etc/php/conf.d, então não precisa habilitar nada à mão aqui.
    instalar_pacotes_aur php-sqlsrv php-pdo_sqlsrv
    sucesso "Extensões PHP configuradas."
}

registrar_modulo "php_extensoes" "Extensões PHP (SQL Server)" \
    "Instala sqlsrv/pdo_sqlsrv (pré-compilados via Chaotic-AUR)" \
    "configurar_extensoes_php" "pacotes_base"
