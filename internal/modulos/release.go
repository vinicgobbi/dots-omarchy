package modulos

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

// assetRelease é um arquivo de uma release do GitHub. SHA256 vem do campo
// "digest" que o GitHub calcula no upload (vazio em releases antigas).
type assetRelease struct {
	Nome   string
	URL    string
	SHA256 string
}

// ultimoAsset devolve o primeiro asset da última release de um repositório do
// GitHub que satisfaça casa. No --dry-run não consulta a rede.
func ultimoAsset(s *sistema.Sistema, repo string, casa func(nome string) bool) (assetRelease, error) {
	api := "https://api.github.com/repos/" + repo + "/releases/latest"
	if s.DryRun {
		s.UI.DryRun("consultar " + api)
		return assetRelease{Nome: "asset de " + repo, URL: "https://github.com/" + repo + "/releases/latest"}, nil
	}
	corpo, err := s.Get(api, 30*time.Second)
	if err != nil {
		return assetRelease{}, err
	}
	var release struct {
		Assets []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.Unmarshal(corpo, &release); err != nil {
		return assetRelease{}, err
	}
	for _, a := range release.Assets {
		if casa(a.Name) {
			sha, _ := strings.CutPrefix(a.Digest, "sha256:")
			return assetRelease{Nome: a.Name, URL: a.URL, SHA256: sha}, nil
		}
	}
	return assetRelease{}, fmt.Errorf("nenhum asset compatível na última release de %s", repo)
}

// baixarAsset baixa um asset conferindo o SHA-256 publicado pelo GitHub, quando
// houver.
func baixarAsset(s *sistema.Sistema, a assetRelease, destino string) error {
	if a.SHA256 == "" {
		if !s.DryRun {
			s.UI.Aviso(a.Nome + " não tem SHA-256 publicado na release: baixando sem conferência de integridade.")
		}
		return s.Baixar(a.URL, destino)
	}
	return s.BaixarVerificado(a.URL, destino, a.SHA256)
}
