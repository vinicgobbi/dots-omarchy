package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

var temasOmarchy = &Modulo{
	ID:         "temas_omarchy",
	Titulo:     "Instalar temas extras",
	Descricao:  "Pergunta antes de instalar os temas listados em data/themes.json (o último fica ativo)",
	Executar:   instalarTemasOmarchy,
	Interativo: true,
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
		Itens:    temas,
		Pergunta: fmt.Sprintf("Instalar os %d temas acima?", len(temas)),
	}) {
		s.UI.Aviso("Temas não instalados (escolha do usuário).")
		return nil
	}

	// O omarchy-theme-install clona e já aplica o tema, então o último da
	// lista é o que fica ativo. Fora da sessão gráfica a aplicação pode
	// falhar, mas o clone fica feito e o tema aparece no seletor.
	for _, url := range temas {
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
