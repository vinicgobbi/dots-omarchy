package modulos

import (
	"fmt"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var rust = &Modulo{
	ID:     "rust",
	Grupo:  GrupoDesenvolvimento,
	Titulo: "Rust, eza, topgrade e cargo-update",
	Descricao: "rustup e compilação de eza, topgrade e cargo-update (cargo install-update, usado pelo " +
		"topgrade para atualizar o que veio do cargo) via cargo",
	Executar: instalarRust,
}

func instalarRust(s *sistema.Sistema) error {
	s.UI.Info("Instalando rustup e compilando eza/topgrade/cargo-update via cargo...")

	// O cargo-update compila com libgit2/libcurl/libssh2 (embutidos se faltarem
	// no sistema), o que exige CMake, pkg-config e os cabeçalhos do OpenSSL.
	if err := s.Pacman(s.Dados.Lista("rust", "pacman")...); err != nil {
		return err
	}
	if err := s.Usr("curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs " +
		"| sh -s -- -y --default-toolchain stable"); err != nil {
		return err
	}
	if err := s.Usr(`"$HOME/.cargo/bin/cargo" install --locked ` +
		strings.Join(s.Dados.Lista("rust", "cargo"), " ")); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("rustup, eza, topgrade e cargo-update instalados para %s.", s.Usuario()))
	return nil
}
