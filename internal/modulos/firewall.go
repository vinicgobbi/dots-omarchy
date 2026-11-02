package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

var firewall = &Modulo{
	ID:        "firewall",
	Grupo:     GrupoRede,
	Titulo:    "Firewall (ufw + gufw)",
	Descricao: "Instala o ufw e a interface gráfica gufw (não mexe no estado nem nas regras do firewall)",
	Executar:  instalarFirewall,
}

func instalarFirewall(s *sistema.Sistema) error {
	s.UI.Info("Instalando o ufw e o gufw...")
	if err := s.Pacman(s.Dados.Lista("firewall", "pacman")...); err != nil {
		return err
	}
	s.UI.Sucesso("ufw e gufw instalados.")
	return nil
}
