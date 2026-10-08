// Package dados lê as listas editáveis do projeto, em data/*.json.
package dados

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Webapp struct {
	Nome  string `json:"nome"`
	URL   string `json:"url"`
	Icone string `json:"icone"`
}

type Dot struct {
	Origem       string `json:"origem"`
	Destino      string `json:"destino"`
	Modo         string `json:"modo"`
	TrocarHome   bool   `json:"trocar_home"`
	Opcional     bool   `json:"opcional"`
	AvisoAusente string `json:"aviso_ausente"`
}

type Dados struct {
	Flatpaks struct {
		Apps  []string `json:"apps"`
		Jogos []string `json:"jogos"`
	}
	Plugins  []string
	Themes   []string
	Webapps  []Webapp
	Packages map[string]map[string]json.RawMessage
	Dots     []Dot
}

// Carregar lê todos os JSON de data/. Um JSON inválido para o setup logo no
// início, antes de qualquer mudança no sistema.
func Carregar(raiz string) (*Dados, error) {
	d := &Dados{}
	arquivos := []struct {
		nome    string
		destino any
	}{
		{"flatpaks", &d.Flatpaks},
		{"plugins", &d.Plugins},
		{"themes", &d.Themes},
		{"webapps", &d.Webapps},
		{"packages", &d.Packages},
		{"dots", &d.Dots},
	}
	for _, a := range arquivos {
		caminho := filepath.Join(raiz, "data", a.nome+".json")
		conteudo, err := os.ReadFile(caminho)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(conteudo, a.destino); err != nil {
			return nil, fmt.Errorf("%s: %w", caminho, err)
		}
	}
	return d, nil
}

// Lista devolve packages.json[modulo][chave] como lista de strings.
func (d *Dados) Lista(modulo, chave string) []string {
	var lista []string
	json.Unmarshal(d.Packages[modulo][chave], &lista)
	return lista
}

// Texto devolve packages.json[modulo][chave] como string.
func (d *Dados) Texto(modulo, chave string) string {
	var texto string
	json.Unmarshal(d.Packages[modulo][chave], &texto)
	return texto
}
