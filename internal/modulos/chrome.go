package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var chrome = &Modulo{
	ID:        "chrome",
	Grupo:     GrupoAplicativos,
	Titulo:    "Google Chrome",
	Descricao: "Instala o Google Chrome integrado ao tema do Omarchy",
	Executar:  instalarChrome,
}

func instalarChrome(s *sistema.Sistema) error {
	s.UI.Info("Instalando o Google Chrome...")
	// omarchy-install-browser instala via yay e usa sudo para criar a pasta de
	// política do Chrome (/etc/opt/chrome/policies/managed) e gravar a cor do
	// tema. Sob `su -c` não há terminal para o sudo pedir senha, o que deixava
	// o Chrome sem o tema; com a regra temporária de sistema.LiberarSudo o sudo
	// não pede nada e o script faz tudo sozinho.
	if err := s.Usr("omarchy-install-browser chrome"); err != nil {
		return err
	}
	s.UI.Sucesso("Google Chrome instalado.")
	return nil
}
