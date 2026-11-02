package modulos

import (
	"fmt"
	"path"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

var temasOmarchy = &Modulo{
	ID:        "temas_omarchy",
	Grupo:     GrupoOmarchy,
	Titulo:    "Temas extras",
	Descricao: "Pergunta antes de instalar os temas listados em data/themes.json e deixa ativo o tema escolhido",
	Executar:  instalarTemasOmarchy,
	Confirma:  true,
}

// instalarTemasOmarchy: temas extras vêm do GitHub de cada autor e podem
// trazer hooks (theme-set.d) que rodam como o usuário a cada troca de tema;
// por isso, como os plugins, só instala depois de confirmar. Sem terminal
// interativo, não instala nada.
func instalarTemasOmarchy(s *sistema.Sistema) error {
	temas := s.Dados.Themes
	if len(temas) == 0 {
		s.UI.Aviso("Nenhum tema listado em data/themes.json; nada a fazer.")
		return nil
	}
	if !s.UI.Interativo() {
		s.UI.Aviso("Sem terminal interativo para confirmar: temas de terceiros NÃO foram instalados.")
		return nil
	}

	if !s.UI.Confirmar(ui.Confirmacao{
		Titulo: "Temas de terceiros",
		Texto: []string{
			"Temas que serão instalados (código de terceiros, podem trazer hooks",
			"que rodam a cada troca de tema). O último da lista fica ativo:",
		},
		Itens:    OrdenarTemas(temas, s.TemaAtivo),
		Pergunta: fmt.Sprintf("Instalar os %d temas acima?", len(temas)),
	}) {
		s.UI.Info("Temas não instalados (escolha do usuário).")
		return nil
	}

	// O omarchy-theme-install clona e já aplica o tema, então o último da
	// lista é o que fica ativo. Fora da sessão gráfica a aplicação pode
	// falhar, mas o clone fica feito e o tema aparece no seletor.
	for _, url := range OrdenarTemas(temas, s.TemaAtivo) {
		s.UI.Info("Tema do Omarchy: " + url)
		ok, err := s.Tentar(sistema.Juntar("omarchy-theme-install", url), sistema.Opts{})
		if err != nil {
			return err
		}
		if !ok {
			s.UI.Aviso(fmt.Sprintf("Não foi possível instalar/aplicar o tema %s; rode depois: omarchy-theme-install %s", url, url))
		}
	}
	s.UI.Sucesso("Temas instalados. Troque pelo seletor de temas do Omarchy, se quiser.")
	return nil
}

// OrdenarTemas põe o tema ativo por último, já que o omarchy-theme-install
// aplica cada tema que instala. Sem escolha (ou com um tema fora da lista), a
// ordem de data/themes.json é mantida.
func OrdenarTemas(temas []string, ativo string) []string {
	ordenados := make([]string, 0, len(temas))
	achou := false
	for _, url := range temas {
		if url == ativo {
			achou = true
			continue
		}
		ordenados = append(ordenados, url)
	}
	if achou {
		ordenados = append(ordenados, ativo)
	}
	return ordenados
}

// NomeTema extrai o nome curto da URL do repositório do tema
// (".../omarchy-lunar-quest-theme" vira "lunar-quest").
func NomeTema(url string) string {
	nome := strings.TrimSuffix(path.Base(url), ".git")
	nome = strings.TrimPrefix(nome, "omarchy-")
	return strings.TrimSuffix(nome, "-theme")
}
