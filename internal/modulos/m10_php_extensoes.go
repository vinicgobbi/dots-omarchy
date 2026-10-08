package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var phpExtensoes = &Modulo{
	ID:        "php_extensoes",
	Titulo:    "Extensões PHP (SQL Server)",
	Descricao: "Instala sqlsrv/pdo_sqlsrv (pré-compilados via Chaotic-AUR)",
	Executar:  configurarExtensoesPHP,
	Deps:      []string{"pacotes_base"},
}

func configurarExtensoesPHP(s *sistema.Sistema) error {
	s.UI.Info("Configurando extensões PHP (sqlsrv)...")

	// O php oficial do Arch não traz pecl/pear; php-sqlsrv/php-pdo_sqlsrv da
	// AUR já vêm pré-compilados (via Chaotic-AUR) e registram o próprio .ini
	// em /etc/php/conf.d, então não precisa habilitar nada à mão aqui.
	if err := s.Aur(s.Dados.Lista("php_extensoes", "aur")...); err != nil {
		return err
	}
	s.UI.Sucesso("Extensões PHP configuradas.")
	return nil
}
