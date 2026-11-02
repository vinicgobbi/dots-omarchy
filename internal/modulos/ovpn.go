package modulos

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var (
	dnsOVPN     = regexp.MustCompile(`(?m)^\s*dhcp-option\s+DNS\s+(\S+)`)
	dominioOVPN = regexp.MustCompile(`(?m)^\s*dhcp-option\s+DOMAIN(?:-SEARCH)?\s+(\S+)`)
)

var ovpn = &Modulo{
	ID:     "ovpn",
	Grupo:  GrupoRede,
	Titulo: "Importar perfis OpenVPN",
	Descricao: "Instala o plugin OpenVPN do NetworkManager e importa os .ovpn de ./OVPN " +
		"(aplicando DNS/domínio quando presentes no arquivo)",
	Executar: importarOVPN,
}

// conexoesNM interpreta a saída de "nmcli -g NAME,TYPE connection show" (nome
// -> tipo). No modo -g, ":" e "\" dentro dos valores vêm escapados com "\".
func conexoesNM(saida string) map[string]string {
	conexoes := map[string]string{}
	for _, linha := range strings.Split(saida, "\n") {
		if linha == "" {
			continue
		}
		var campos []string
		var atual strings.Builder
		for i := 0; i < len(linha); i++ {
			switch {
			case linha[i] == '\\' && i+1 < len(linha):
				i++
				atual.WriteByte(linha[i])
			case linha[i] == ':':
				campos = append(campos, atual.String())
				atual.Reset()
			default:
				atual.WriteByte(linha[i])
			}
		}
		campos = append(campos, atual.String())
		if len(campos) == 2 {
			conexoes[campos[0]] = campos[1]
		}
	}
	return conexoes
}

// avisarPermissaoOVPN avisa quando um .ovpn (que costuma trazer chave privada e
// certificados embutidos) pode ser lido por qualquer usuário da máquina.
func avisarPermissaoOVPN(s *sistema.Sistema, arquivo string) {
	if info, err := os.Stat(arquivo); err == nil && info.Mode().Perm()&0o004 != 0 {
		s.UI.Aviso(fmt.Sprintf("%s pode ser lido por qualquer usuário (permissão %o) e costuma conter chaves "+
			"privadas: considere 'chmod 600'.", filepath.Base(arquivo), info.Mode().Perm()))
	}
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

	res, err := s.Rodar([]string{"nmcli", "-g", "NAME,TYPE", "connection", "show"},
		sistema.Opts{Capturar: true, Leitura: true, SemCheck: true})
	if err != nil {
		return err
	}
	existentes := conexoesNM(res.Saida)

	for _, arquivo := range arquivos {
		nome := strings.TrimSuffix(filepath.Base(arquivo), ".ovpn")

		avisarPermissaoOVPN(s, arquivo)

		// Reimporta do zero para continuar idempotente em reexecuções, mas só
		// apaga se a conexão com esse nome for mesmo uma VPN, para não levar
		// junto uma conexão cabeada/Wi-Fi homônima.
		if tipo, ok := existentes[nome]; ok {
			if tipo != "vpn" {
				s.UI.Aviso(fmt.Sprintf("Já existe uma conexão '%s' do tipo %s (não VPN): pulando %s para não sobrescrevê-la.",
					nome, tipo, filepath.Base(arquivo)))
				continue
			}
			if _, err := s.Rodar([]string{"nmcli", "connection", "delete", "id", nome}, sistema.Opts{Quieto: true}); err != nil {
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
