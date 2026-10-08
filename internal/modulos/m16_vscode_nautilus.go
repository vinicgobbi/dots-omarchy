package modulos

import (
	"fmt"
	"path"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var vscodeNautilus = &Modulo{
	ID:        "vscode_nautilus",
	Titulo:    "Extensões do Nautilus (VSCode, copiar caminho e abrir no terminal)",
	Descricao: "Adiciona 'Abrir com o VSCode', 'Copiar caminho' e 'Abrir no Terminal' ao menu de contexto do Nautilus",
	Executar:  instalarVSCodeNautilus,
	Deps:      []string{"pacotes_base"},
}

func instalarVSCodeNautilus(s *sistema.Sistema) error {
	s.UI.Info("Instalando extensões do VSCode, de copiar caminho e de abrir no terminal para o Nautilus...")

	if err := s.Pacman(s.Dados.Lista("vscode_nautilus", "pacman")...); err != nil {
		return err
	}

	// Cada extensão é um repositório com install.sh próprio, rodado como o
	// usuário-alvo a partir de um clone em /tmp.
	for _, repo := range s.Dados.Lista("vscode_nautilus", "extensoes_git") {
		base := path.Base(repo)
		pasta := "/tmp/" + base[:len(base)-len(path.Ext(base))]
		ok, err := s.Tentar(fmt.Sprintf("rm -rf %s && git clone %s %s && bash %s/install.sh",
			pasta, repo, pasta, pasta), sistema.Opts{})
		if err != nil {
			return err
		}
		if !ok {
			s.UI.Aviso(fmt.Sprintf("Falha ao instalar a extensão do Nautilus de %s.", repo))
		}
	}
	s.UI.Sucesso("Extensões do VSCode, de copiar caminho e de abrir no terminal instaladas no Nautilus.")
	return nil
}
