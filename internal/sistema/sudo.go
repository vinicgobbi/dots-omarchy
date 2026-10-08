package sistema

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// SudoersTemporario libera sudo sem senha para o usuário-alvo enquanto o
// setup roda. Os scripts do Omarchy (omarchy-update, omarchy-install-browser,
// omarchy-theme-set-browser…) e o yay chamam sudo por conta própria, e sob
// `su -c` não há terminal para o sudo pedir a senha: sem isso eles falham ou
// tiram a tela da TUI. O setup já roda como root (a senha foi pedida no
// `sudo ./setup.sh`), então a regra não dá ao usuário nada que ele já não
// tenha; ela só existe durante a execução e é removida ao fim, mesmo em falha
// ou abort.
const SudoersTemporario = "/etc/sudoers.d/99-post-omarchy"

// sudoersLegado é a regra antiga, só do pacman (versões anteriores do setup).
const sudoersLegado = "/etc/sudoers.d/99-post-omarchy-aur"

// LiberarSudo cria a regra temporária e devolve a função que a remove.
func (s *Sistema) LiberarSudo() (func(), error) {
	if s.DryRun {
		s.UI.DryRun(fmt.Sprintf("escrever %s (sudo sem senha para %s durante o setup)", SudoersTemporario, s.Usuario()))
		return func() { s.UI.DryRun("rm -f " + SudoersTemporario) }, nil
	}
	conteudo := fmt.Sprintf("# Temporário: criado e removido pelo post-omarchy.\n%s ALL=(ALL) NOPASSWD: ALL\n", s.Usuario())
	if err := os.WriteFile(SudoersTemporario, []byte(conteudo), 0o440); err != nil {
		return nil, err
	}
	if err := os.Chmod(SudoersTemporario, 0o440); err != nil {
		RemoverSudo()
		return nil, err
	}
	if err := exec.Command("visudo", "-cqf", SudoersTemporario).Run(); err != nil {
		RemoverSudo()
		return nil, errors.New("falha ao validar a regra temporária de sudo")
	}
	s.Log.Linha("INFO", "Regra temporária de sudo criada: "+SudoersTemporario)
	return func() {
		RemoverSudo()
		s.Log.Linha("INFO", "Regra temporária de sudo removida.")
	}, nil
}

// RemoverSudo apaga a regra temporária (e a legada). Idempotente: é chamada
// também na saída do programa e em sinais, para a regra nunca ficar para trás.
func RemoverSudo() {
	if os.Geteuid() != 0 {
		return
	}
	os.Remove(SudoersTemporario)
	os.Remove(sudoersLegado)
}
