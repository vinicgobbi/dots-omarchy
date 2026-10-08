package modulos

import (
	"fmt"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
	"github.com/vinicgobbi/dots-omarchy/internal/ui"
)

var pluginsOmarchy = &Modulo{
	ID:         "plugins_omarchy",
	Titulo:     "Instalar plugins de terceiros",
	Descricao:  "Mostra os riscos e pergunta antes de instalar os plugins da barra listados em data/plugins.json",
	Executar:   instalarPluginsOmarchy,
	Interativo: true,
}

var riscosPlugins = []string{
	"Os plugins abaixo são baixados direto do GitHub de cada autor e NÃO",
	"são revisados nem mantidos pelo Omarchy nem por este script.",
	"",
	"• Rodam como código arbitrário, sem sandbox, dentro do omarchy-shell,",
	"  com as mesmas permissões do seu usuário: podem ler e alterar seus",
	"  arquivos, executar comandos e acessar a rede.",
	"• É instalada a versão atual de cada repositório, que pode ter mudado",
	"  (ou sido comprometida) desde que esta lista foi montada.",
	"• Um plugin com bug pode travar ou deixar lenta a barra e o shell.",
	"• Alguns pedem passos extras (dependências, scripts de instalação);",
	"  veja o README de cada um.",
	"",
	"Revise o código antes, se puder. Plugins que serão instalados:",
}

// instalarPluginsOmarchy: plugins são código de terceiros que roda dentro do
// omarchy-shell: por isso este é um dos últimos módulos e só instala depois de
// mostrar os riscos e o usuário confirmar. Sem terminal interativo, não
// instala nada.
func instalarPluginsOmarchy(s *sistema.Sistema) error {
	plugins := s.Dados.Plugins
	if len(plugins) == 0 {
		s.UI.Aviso("Nenhum plugin listado em data/plugins.json; nada a fazer.")
		return nil
	}
	if !s.UI.Interativo() {
		s.UI.Aviso("Sem terminal interativo para confirmar: plugins de terceiros NÃO foram instalados.")
		return nil
	}

	if !s.UI.Confirmar(ui.Confirmacao{
		Titulo:   "ATENÇÃO: plugins de terceiros",
		Texto:    riscosPlugins,
		Itens:    plugins,
		Pergunta: fmt.Sprintf("Entendi os riscos. Instalar os %d plugins acima?", len(plugins)),
		Perigo:   true,
	}) {
		s.UI.Aviso("Plugins de terceiros não instalados (escolha do usuário).")
		return nil
	}

	// --yes porque a confirmação já foi feita acima (e o su não tem o gum
	// interativo do omarchy). Cada plugin é best-effort: precisa de rede e
	// falha se já estiver instalado.
	for _, url := range plugins {
		s.UI.Info("Plugin do Omarchy: " + url)
		ok, err := s.Tentar(sistema.Juntar("omarchy", "plugin", "add", url, "--yes"), sistema.Opts{})
		if err != nil {
			return err
		}
		if !ok {
			s.UI.Aviso(fmt.Sprintf("Não foi possível instalar o plugin %s; rode depois: omarchy plugin add %s", url, url))
		}
	}
	s.UI.Sucesso("Plugins instalados. A posição deles na barra vem do shell.json dos dots.")
	return nil
}
