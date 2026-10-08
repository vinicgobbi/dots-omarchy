package modulos

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

const (
	gcmReleaseAPI = "https://api.github.com/repos/git-ecosystem/git-credential-manager/releases/latest"
	gcmTar        = "/tmp/gcm.tar.gz"
	gcmDir        = "/usr/local/gcm"
)

var chromeGCM = &Modulo{
	ID:         "chrome_gcm",
	Titulo:     "Chrome + Git Credential Manager",
	Descricao:  "Instala o Google Chrome (integrado ao tema do Omarchy) e o Git Credential Manager",
	Executar:   instalarChromeEGCM,
	Interativo: true,
}

// urlGCM devolve o asset .tar.gz linux-x64 (sem os símbolos) da última release.
func urlGCM() (string, error) {
	resposta, err := http.Get(gcmReleaseAPI)
	if err != nil {
		return "", err
	}
	defer resposta.Body.Close()
	var release struct {
		Assets []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resposta.Body).Decode(&release); err != nil {
		return "", err
	}
	for _, a := range release.Assets {
		if strings.HasSuffix(a.Name, ".tar.gz") && strings.Contains(a.Name, "linux-x64") &&
			!strings.Contains(a.Name, "symbols") {
			return a.URL, nil
		}
	}
	return "", errors.New("release do Git Credential Manager sem o .tar.gz linux-x64")
}

func instalarChromeEGCM(s *sistema.Sistema) error {
	s.UI.Info("Instalando Git Credential Manager e Google Chrome...")

	// GCM (tar.gz local) e Chrome são independentes: o download do GCM roda
	// em segundo plano enquanto o Chrome é instalado.
	url, err := urlGCM()
	if err != nil {
		return err
	}
	download := make(chan error, 1)
	go func() { download <- s.Run("curl", "-sSL", "-o", gcmTar, url) }()

	// omarchy-install-browser instala via yay e tenta criar a pasta de política
	// do Chrome com sudo. Sob `su -c` o sudo pode não conseguir pedir a senha
	// ("conversation failed") e o script segue sem a pasta, então o tema nunca
	// chega ao Chrome. Como aqui já somos root, criamos a pasta com o próprio
	// helper do Omarchy e reaplicamos o tema como o usuário.
	_, errChrome := s.ComoUsuario("omarchy-install-browser chrome", sistema.Opts{Interativo: true})
	if err := errors.Join(errChrome, <-download); err != nil {
		return err
	}
	if err := s.Run("bash", "-c", `source "${OMARCHY_PATH:-/usr/share/omarchy}/install/helpers/browser-policy.sh" `+
		`&& browser_policy_setup_dir /etc/opt/chrome/policies/managed`); err != nil {
		return err
	}
	if err := s.Usr("omarchy-theme-set-browser"); err != nil {
		return err
	}

	if err := s.Run("mkdir", "-p", gcmDir); err != nil {
		return err
	}
	if err := s.Run("tar", "-xzf", gcmTar, "-C", gcmDir); err != nil {
		return err
	}
	if err := s.Link(filepath.Join(gcmDir, "git-credential-manager"), "/usr/local/bin/git-credential-manager"); err != nil {
		return err
	}
	s.UI.Sucesso("Chrome e GCM instalados.")
	return nil
}
