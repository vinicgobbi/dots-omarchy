package modulos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var (
	dnsOVPN     = regexp.MustCompile(`(?m)^\s*dhcp-option\s+DNS\s+(\S+)`)
	dominioOVPN = regexp.MustCompile(`(?m)^\s*dhcp-option\s+DOMAIN(?:-SEARCH)?\s+(\S+)`)
)

var ovpn = &Modulo{
	ID:     "ovpn",
	Titulo: "Importar perfis OpenVPN",
	Descricao: "Instala o plugin OpenVPN do NetworkManager e importa os .ovpn de ./OVPN " +
		"(aplicando DNS/domínio quando presentes no arquivo)",
	Executar: importarOVPN,
}

func capturas(re *regexp.Regexp, texto string) string {
	var valores []string
	for _, m := range re.FindAllStringSubmatch(texto, -1) {
		valores = append(valores, m[1])
	}
	return strings.Join(valores, " ")
}

func importarOVPN(s *sistema.Sistema) error {
	s.UI.Info("Importando perfis OpenVPN para o NetworkManager...")

	plugin := s.Dados.Lista("ovpn", "pacman")
	if !s.Ok(append([]string{"pacman", "-Qq"}, plugin...)...) {
		if err := s.Pacman(plugin...); err != nil {
			return err
		}
	}

	pasta := filepath.Join(s.Raiz, "OVPN")
	arquivos, _ := filepath.Glob(filepath.Join(pasta, "*.ovpn"))
	if len(arquivos) == 0 {
		s.UI.Aviso(fmt.Sprintf("Nenhum arquivo .ovpn encontrado em %s — pulando importação.", pasta))
		return nil
	}

	res, err := s.Rodar([]string{"nmcli", "-g", "NAME", "connection", "show"},
		sistema.Opts{Capturar: true, Leitura: true})
	if err != nil {
		return err
	}
	existentes := strings.Split(res.Saida, "\n")

	for _, arquivo := range arquivos {
		nome := strings.TrimSuffix(filepath.Base(arquivo), ".ovpn")

		// Reimportar do zero para o script continuar idempotente em reexecuções.
		if slices.Contains(existentes, nome) {
			if _, err := s.Rodar([]string{"nmcli", "connection", "delete", nome}, sistema.Opts{Quieto: true}); err != nil {
				return err
			}
		}

		res, err := s.Rodar([]string{"nmcli", "connection", "import", "type", "openvpn", "file", arquivo},
			sistema.Opts{SemCheck: true, Quieto: true})
		if err != nil {
			return err
		}
		if res.Codigo != 0 {
			s.UI.Aviso(fmt.Sprintf("Falha ao importar %s, pulando.", arquivo))
			continue
		}

		// DNS e domínio de busca não vêm sempre preenchidos pela importação
		// automática do plugin; quando o .ovpn declara essas diretivas, aplica
		// manualmente via nmcli.
		conteudo, err := os.ReadFile(arquivo)
		if err != nil {
			return err
		}
		if dns := capturas(dnsOVPN, string(conteudo)); dns != "" {
			if err := s.Run("nmcli", "connection", "modify", nome, "ipv4.dns", dns); err != nil {
				return err
			}
		}
		if dominios := capturas(dominioOVPN, string(conteudo)); dominios != "" {
			if err := s.Run("nmcli", "connection", "modify", nome, "ipv4.dns-search", dominios); err != nil {
				return err
			}
		}
		s.UI.Sucesso(fmt.Sprintf("Perfil '%s' importado.", nome))
	}
	return nil
}
