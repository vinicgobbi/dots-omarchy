package modulos

import (
	"fmt"
	"strings"

	"github.com/vinicgobbi/dots-omarchy/internal/sistema"
)

var rustTools = &Modulo{
	ID:        "rust_tools",
	Titulo:    "Instalar Rust, eza e topgrade",
	Descricao: "rustup e compilação de eza e topgrade via cargo",
	Executar:  instalarRustTools,
}

func instalarRustTools(s *sistema.Sistema) error {
	s.UI.Info("Instalando rustup e compilando eza/topgrade via cargo...")

	if err := s.Pacman(s.Dados.Lista("rust_tools", "pacman")...); err != nil {
		return err
	}
	if err := s.Usr("curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs " +
		"| sh -s -- -y --default-toolchain stable"); err != nil {
		return err
	}
	if err := s.Usr(`"$HOME/.cargo/bin/cargo" install --locked ` +
		strings.Join(s.Dados.Lista("rust_tools", "cargo"), " ")); err != nil {
		return err
	}
	s.UI.Sucesso(fmt.Sprintf("rustup, eza e topgrade instalados para %s.", s.Usuario()))
	return nil
}
