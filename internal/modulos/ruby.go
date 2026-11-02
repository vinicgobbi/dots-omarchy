package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var ruby = &Modulo{
	ID:     "ruby",
	Grupo:  GrupoDesenvolvimento,
	Titulo: "Ruby, Bundler e Jekyll",
	Descricao: "Ruby e, como gems do usuário (sem sudo), Bundler e Jekyll; o Bundler instala as gems de " +
		"cada projeto em vendor/bundle",
	Executar: instalarRuby,
}

// O Bundler e o Jekyll são instalados como gems do usuário (GEM_HOME na home),
// como recomenda a documentação do Jekyll: assim gem/bundle nunca pedem sudo.
const rubyAmbiente = `# Gems do Ruby na home do usuário (gem/bundle sem sudo) — post-omarchy
export GEM_HOME="$HOME/.local/share/gem"
export PATH="$GEM_HOME/bin:$PATH"
`

// scriptRuby instala as gems na home do usuário, configura o Bundler e
// acrescenta GEM_HOME/PATH ao .zshrc e ao .bashrc que existirem (uma vez só).
// "path vendor/bundle" (global, em ~/.bundle/config): o bundle install de cada
// projeto grava as gems dentro do próprio projeto.
func scriptRuby(gems []string) string {
	return "set -e\n" + rubyAmbiente + "\n" +
		sistema.Juntar(append([]string{"gem", "install", "--no-document"}, gems...)...) + "\n" +
		`bundle config set --global path vendor/bundle

for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
    [ -f "$rc" ] || continue
    grep -qF 'GEM_HOME=' "$rc" || printf '\n%s' ` + sistema.Juntar(rubyAmbiente) + ` >> "$rc"
done
`
}

func instalarRuby(s *sistema.Sistema) error {
	s.UI.Info("Instalando Ruby e compiladores para gems nativas...")
	if err := s.Pacman(s.Dados.Lista("ruby", "pacman")...); err != nil {
		return err
	}
	s.UI.Info(fmt.Sprintf("Instalando Bundler e Jekyll como gems de %s...", s.Usuario()))
	if err := s.Usr(scriptRuby(s.Dados.Lista("ruby", "gem"))); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("Ruby, Bundler e Jekyll instalados para %s (gems em ~/.local/share/gem; "+
		"projetos em vendor/bundle).", s.Usuario()))
	return nil
}
