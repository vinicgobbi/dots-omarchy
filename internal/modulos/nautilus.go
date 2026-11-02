package modulos

import (
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var nautilus = &Modulo{
	ID:     "nautilus",
	Grupo:  GrupoOmarchy,
	Titulo: "Nautilus: VSCode, copiar caminho, terminal e Projetos",
	Descricao: "Adiciona 'Abrir com o VSCode', 'Copiar caminho' e 'Abrir no Terminal' ao menu de contexto do " +
		"Nautilus e cria a pasta Projetos (diretório XDG e favorito na barra lateral)",
	Executar: instalarNautilus,
}

// scriptExtensoesNautilus clona cada extensão no commit fixado e roda o
// install.sh dela, como o usuário-alvo.
func scriptExtensoesNautilus() string {
	extensoes := []struct{ nome, repo, commit string }{
		{"vscode_nautilus", vscodeNautilusRepo, vscodeNautilusCommit},
		{"nautilus-copypath", copypathRepo, copypathCommit},
		{"openinterminal", abrirTerminalRepo, abrirTerminalCommit},
	}
	script := inicioScriptTemp
	for _, e := range extensoes {
		pasta := `"$dir/` + e.nome + `"`
		script += clonarFixado(e.repo, e.commit, pasta) + `bash "$dir/` + e.nome + `/install.sh"` + "\n"
	}
	return script
}

func instalarNautilus(s *sistema.Sistema) error {
	s.UI.Info("Instalando extensões do VSCode, de copiar caminho e de abrir no terminal para o Nautilus...")
	if err := s.Pacman(s.Dados.Lista("nautilus", "pacman")...); err != nil {
		return err
	}
	if err := s.Usr(scriptExtensoesNautilus()); err != nil {
		return err
	}
	if err := configurarProjetos(s); err != nil {
		return err
	}
	s.UI.Sucesso("Extensões do Nautilus instaladas e pasta de projetos criada.")
	return nil
}

// configurarProjetos cria o diretório de Projetos (XDG e bookmark do
// Nautilus), com o nome no idioma do usuário (o LANG vem do login shell dele).
func configurarProjetos(s *sistema.Sistema) error {
	res, err := s.ComoUsuario(`printf %s "$LANG"`, sistema.Opts{Capturar: true, Leitura: true, SemCheck: true})
	if err != nil {
		return err
	}
	nome := "Projects"
	if strings.HasPrefix(res.Saida, "pt_") {
		nome = "Projetos"
	}
	projetos := filepath.Join(s.Home(), nome)
	if err := s.CriarDirUsuario(projetos); err != nil {
		return err
	}
	if err := s.Usr(sistema.Juntar("xdg-user-dirs-update", "--set", "PROJECTS", projetos)); err != nil {
		return err
	}

	bookmarks := filepath.Join(s.Home(), ".config/gtk-3.0/bookmarks")
	if !contem(bookmarks, "file://"+projetos) {
		return s.Escrever(bookmarks, "file://"+projetos+"\n", sistema.Escrita{Anexar: true, DoUsuario: true})
	}
	return nil
}
