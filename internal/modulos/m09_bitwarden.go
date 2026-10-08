package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var bitwarden = &Modulo{
	ID:        "bitwarden",
	Titulo:    "Instalar Bitwarden nativo",
	Descricao: "Instala o Bitwarden desktop (necessário para integração nativa com o navegador)",
	Executar:  instalarBitwardenNativo,
	Deps:      []string{"repositorios"},
}

func instalarBitwardenNativo(s *sistema.Sistema) error {
	s.UI.Info("Instalando Bitwarden nativo (integração com navegador)...")
	if err := s.Aur(s.Dados.Lista("bitwarden", "aur")...); err != nil {
		return err
	}
	s.UI.Sucesso("Bitwarden instalado nativamente.")
	return nil
}
