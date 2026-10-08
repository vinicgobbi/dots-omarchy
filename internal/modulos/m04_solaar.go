package modulos

import "github.com/vinicgobbi/dots-omarchy/internal/sistema"

const regrasUdevURL = "https://raw.githubusercontent.com/pwr-Solaar/Solaar/master/" +
	"rules.d-uinput/42-logitech-unify-permissions.rules"

var solaar = &Modulo{
	ID:        "solaar",
	Titulo:    "Instalar e configurar Solaar",
	Descricao: "Instala o Solaar (nativo, com fallback para Flatpak) e as regras UDEV do receptor Logitech Unifying",
	Executar:  configurarSolaar,
	Deps:      []string{"pacotes_base"},
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

	s.UI.Info("Configurando regras UDEV para o Solaar...")
	for _, passo := range [][]string{
		{"curl", "-sL", regrasUdevURL, "-o", "/etc/udev/rules.d/42-logitech-unify-permissions.rules"},
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
