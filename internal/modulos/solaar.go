package modulos

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var solaar = &Modulo{
	ID:     "solaar",
	Grupo:  GrupoAplicativos,
	Titulo: "Solaar (Logitech)",
	Descricao: "Instala o Solaar (nativo, com fallback para Flatpak), as regras UDEV do receptor Logitech " +
		"Unifying e o início junto da sessão, só na bandeja",
	Executar: configurarSolaar,
}

// scriptAutostartSolaar põe o Solaar (pacote, ou Flatpak como alternativa) para
// iniciar com a sessão só no ícone da bandeja ("-w hide"), sem abrir a janela.
const scriptAutostartSolaar = `mkdir -p "$HOME/.config/autostart"
for desktop in /usr/share/applications/solaar.desktop \
    /var/lib/flatpak/exports/share/applications/io.github.pwr_solaar.solaar.desktop; do
    if [ -f "$desktop" ]; then
        destino="$HOME/.config/autostart/$(basename "$desktop")"
        cp "$desktop" "$destino"
        sed -i -E '/^Exec=/{/ -w hide$/! s/$/ -w hide/}' "$destino"
        break
    fi
done
`

const regraSolaarNome = "42-logitech-unify-permissions.rules"

// solaarTag é a versão do Solaar de onde veio a regra embutida abaixo
// (rules.d-uinput/42-logitech-unify-permissions.rules). Ela vem no binário em
// vez de baixada: regras UDEV rodam como root (podem ter RUN+=), então não
// convém buscá-las de uma branch que muda. O workflow pins.yml atualiza as
// duas coisas juntas quando sai versão nova.
const solaarTag = "1.1.20"

//go:embed 42-logitech-unify-permissions.rules
var regraSolaar string

// regraSolaarInstalada diz se algum pacote já instalou a regra do Solaar.
func regraSolaarInstalada() bool {
	for _, pasta := range []string{"/usr/lib/udev/rules.d", "/lib/udev/rules.d"} {
		if _, err := os.Stat(filepath.Join(pasta, regraSolaarNome)); err == nil {
			return true
		}
	}
	return false
}

func configurarSolaar(s *sistema.Sistema) error {
	s.UI.Info("Instalando Solaar...")
	nativo := s.Dados.Lista("solaar", "pacman")

	if s.Ok(append([]string{"pacman", "-Si"}, nativo...)...) {
		if err := s.Pacman(nativo...); err != nil {
			return err
		}
		s.UI.Sucesso("Solaar instalado via pacote nativo (pacman).")
	} else {
		s.UI.Aviso("Pacote nativo do Solaar não encontrado; instalando via Flatpak.")
		if err := s.Flatpak(s.Dados.Texto("solaar", "flatpak_alternativo")); err != nil {
			return err
		}
		s.UI.Sucesso("Solaar instalado via Flatpak.")
	}

	passos := [][]string{{"usermod", "-aG", "plugdev", s.Usuario()}}
	if !sistema.GrupoExiste("plugdev") {
		passos = append([][]string{{"groupadd", "plugdev"}}, passos...)
	}
	for _, passo := range passos {
		if err := s.Run(passo...); err != nil {
			return err
		}
	}

	ok, err := s.Tentar(scriptAutostartSolaar, sistema.Opts{})
	if err != nil {
		return err
	}
	if !ok {
		s.UI.Aviso("Não foi possível pôr o Solaar para iniciar com a sessão.")
	}

	// O pacote nativo costuma trazer a regra em /usr/lib/udev/rules.d; só sem
	// ela (ex.: Flatpak, que não instala nada fora do sandbox) gravamos a nossa.
	if regraSolaarInstalada() {
		s.UI.Info("Regra UDEV do Solaar já instalada pelo pacote.")
		return nil
	}
	s.UI.Info("Configurando regras UDEV para o Solaar (regra do Solaar " + solaarTag + ")...")
	if err := s.Escrever("/etc/udev/rules.d/"+regraSolaarNome, regraSolaar, sistema.Escrita{}); err != nil {
		return err
	}
	for _, passo := range [][]string{
		{"udevadm", "control", "--reload-rules"},
		{"udevadm", "trigger", "--subsystem-match=usb"},
		{"udevadm", "trigger", "--subsystem-match=hidraw"},
	} {
		if err := s.Run(passo...); err != nil {
			return err
		}
	}
	s.UI.Sucesso("Regras UDEV do Solaar configuradas.")
	return nil
}
