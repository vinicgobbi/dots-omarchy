package modulos

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/vinicgobbi/dots-omarchy/internal/dados"
	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

// Arquivos que só existem para manter a pasta no git.
var ignorados = map[string]bool{".gitkeep": true}

var dotsOmarchy = &Modulo{
	ID:     "dots_omarchy",
	Titulo: "Aplicar dots do Omarchy",
	Descricao: "Copia a config do Hyprland, do omarchy-shell, o screensaver, o Solaar, os scripts de " +
		"~/.local/bin e o CLAUDE.md global, se houver (com backup do que existia)",
	Executar: aplicarDotsOmarchy,
}

// instalarDot copia um arquivo de dots/ para ~/<destino> como o usuário-alvo,
// guardando o original em <arquivo>.bak-post-omarchy na primeira vez que ele
// difere do que vamos instalar (para nunca perder customização anterior).
func instalarDot(s *sistema.Sistema, origem string, conteudo []byte, destino string, modo os.FileMode) error {
	alvo := filepath.Join(s.Home(), destino)
	if info, err := os.Stat(alvo); err == nil && info.Mode().IsRegular() {
		atual, err := os.ReadFile(alvo)
		if err != nil {
			return err
		}
		if bytes.Equal(atual, conteudo) {
			return nil
		}
		backup := alvo + ".bak-post-omarchy"
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			if err := s.CopiarPreservando(alvo, backup); err != nil {
				return err
			}
		}
	}

	if err := s.InstalarConteudo(origem, conteudo, alvo, modo); err != nil {
		return err
	}
	s.UI.Info("Instalado ~/" + destino)
	return nil
}

type parDot struct{ origem, destino string }

// arquivosDaEntrada devolve (origem, destino) de uma entrada do dots.json. Com
// glob, o destino é uma pasta e cada arquivo mantém o caminho relativo à parte
// fixa do padrão ("hypr/*.lua" -> ".config/hypr/<nome>.lua").
func arquivosDaEntrada(dots string, e dados.Dot) ([]parDot, error) {
	if !strings.Contains(e.Origem, "*") {
		return []parDot{{filepath.Join(dots, e.Origem), e.Destino}}, nil
	}

	base := filepath.Join(dots, strings.SplitN(e.Origem, "*", 2)[0])
	encontrados, err := doublestar.FilepathGlob(filepath.Join(dots, e.Origem))
	if err != nil {
		return nil, err
	}
	var pares []parDot
	for _, arquivo := range encontrados {
		info, err := os.Stat(arquivo)
		if err != nil || !info.Mode().IsRegular() || ignorados[filepath.Base(arquivo)] {
			continue
		}
		relativo, err := filepath.Rel(base, arquivo)
		if err != nil {
			return nil, err
		}
		pares = append(pares, parDot{arquivo, e.Destino + relativo})
	}
	return pares, nil
}

func aplicarDotsOmarchy(s *sistema.Sistema) error {
	s.UI.Info(fmt.Sprintf("Aplicando dots do Omarchy para %s...", s.Usuario()))
	dots := filepath.Join(s.Raiz, "dots")

	for _, e := range s.Dados.Dots {
		modoTexto := e.Modo
		if modoTexto == "" {
			modoTexto = "0644"
		}
		modo, err := strconv.ParseUint(modoTexto, 8, 32)
		if err != nil {
			return fmt.Errorf("modo inválido em dots.json (%s): %w", e.Origem, err)
		}
		pares, err := arquivosDaEntrada(dots, e)
		if err != nil {
			return err
		}
		for _, p := range pares {
			conteudo, err := os.ReadFile(p.origem)
			if err != nil {
				// Ex.: o CLAUDE.md global não é versionado (ver
				// dots/claude/README.md), então só é instalado se alguém
				// tiver colocado o arquivo lá.
				if e.Opcional {
					aviso := e.AvisoAusente
					if aviso == "" {
						aviso = fmt.Sprintf("dots/%s não encontrado; pulando.", e.Origem)
					}
					s.UI.Aviso(aviso)
					continue
				}
				return fmt.Errorf("dots/%s não encontrado", e.Origem)
			}

			if e.TrocarHome {
				// O Execute do Solaar não expande ~, então o rules.yaml traz o
				// caminho absoluto de /home/vinicius, trocado aqui pelo home do
				// usuário-alvo.
				conteudo = bytes.ReplaceAll(conteudo, []byte("/home/vinicius/"), []byte(s.Home()+"/"))
			}
			if err := instalarDot(s, p.origem, conteudo, p.destino, os.FileMode(modo)); err != nil {
				return err
			}
		}
	}

	s.UI.Sucesso("Dots do Omarchy aplicados. Recarregue o Hyprland (super+shift+r, ou faça logout/login) " +
		"para ver as mudanças.")
	return nil
}
