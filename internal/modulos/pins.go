package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

// Os repositórios pessoais executados pelo setup ficam fixados num commit: uma
// conta comprometida ou um push errado na branch principal não chega sozinho
// às próximas instalações. Para atualizar, troque o commit aqui (o
// scripts/atualizar-pins.sh faz isso).
const (
	dotfilesRepo   = "https://github.com/vinicgobbi/Dotfiles.git"
	dotfilesCommit = "1e9a97699b70582f87e768f1fe6aa1dde9b58fd6"

	vscodeNautilusRepo   = "https://github.com/vinicgobbi/vscode_nautilus.git"
	vscodeNautilusCommit = "77ed67c7b420a3a148c83ce8721f18b9e737cd44"

	copypathRepo   = "https://github.com/vinicgobbi/nautilus-copypath.git"
	copypathCommit = "f2e1c9d0a3b5ed2584539dd9d127cf6121fba599"

	abrirTerminalRepo   = "https://github.com/vinicgobbi/omarchy-nautilus-openinterminal.git"
	abrirTerminalCommit = "1b58b80571f38a28e1dac9cc2888afa24470f871"
)

// inicioScriptTemp abre um script que aborta no primeiro erro e trabalha num
// diretório temporário exclusivo ($dir), apagado ao sair.
const inicioScriptTemp = "set -e\ndir=$(mktemp -d)\ntrap 'rm -rf \"$dir\"' EXIT\n"

// clonarFixado devolve as linhas de shell que clonam repo em dir (expressão de
// shell, ex.: "$dir/x") e posicionam no commit ref.
func clonarFixado(repo, ref, dir string) string {
	return "git clone --quiet " + sistema.Juntar(repo) + " " + dir + "\n" +
		"git -C " + dir + " -c advice.detachedHead=false checkout --quiet " + sistema.Juntar(ref) + "\n"
}
