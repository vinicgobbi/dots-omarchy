package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var phpSQLServer = &Modulo{
	ID:     "php_sqlserver",
	Grupo:  GrupoDesenvolvimento,
	Titulo: "PHP e SQL Server",
	Descricao: "PHP, Composer, unixODBC, as ferramentas do SQL Server e as extensões sqlsrv/pdo_sqlsrv " +
		"(pré-compiladas via Chaotic-AUR)",
	Executar: instalarPHPSQLServer,
	Deps:     []string{"repositorios"},
}

func instalarPHPSQLServer(s *sistema.Sistema) error {
	s.UI.Info("Instalando PHP, Composer e as ferramentas do SQL Server...")
	// PHP/composer e unixODBC são oficiais; msodbcsql, mssql-tools e as
	// extensões só existem na AUR (resolvidos pelo Chaotic-AUR, sem compilar).
	// O php oficial do Arch não traz pecl/pear, mas php-sqlsrv/php-pdo_sqlsrv já
	// registram o próprio .ini em /etc/php/conf.d: não precisa habilitar à mão.
	if err := s.Pacman(s.Dados.Lista("php_sqlserver", "pacman")...); err != nil {
		return err
	}
	if err := s.Aur(s.Dados.Lista("php_sqlserver", "aur")...); err != nil {
		return err
	}
	// O pacote da AUR instala em /opt/mssql-tools (sem versão).
	if err := s.Escrever("/etc/profile.d/mssql-tools.sh", "export PATH=\"$PATH:/opt/mssql-tools/bin\"\n",
		sistema.Escrita{}); err != nil {
		return err
	}
	s.UI.Sucesso("PHP, Composer, ferramentas e extensões do SQL Server instalados.")
	return nil
}
