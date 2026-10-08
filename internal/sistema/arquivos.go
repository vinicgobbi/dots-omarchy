package sistema

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// Escrita ajusta Escrever.
type Escrita struct {
	Modo      os.FileMode // permissão de um arquivo novo (padrão 0644)
	DoUsuario bool        // dono passa a ser o usuário-alvo, como se ele tivesse criado
	Anexar    bool
}

func (s *Sistema) mostrar(acao, caminho string) bool {
	if s.DryRun {
		s.UI.DryRun(acao + " " + caminho)
	}
	return s.DryRun
}

// CriarDirUsuario é o mkdir -p como o usuário: cada diretório criado fica com
// o dono certo.
func (s *Sistema) CriarDirUsuario(caminho string) error {
	if s.mostrar("mkdir -p (usuário)", caminho) {
		return nil
	}
	var faltando []string
	for atual := caminho; ; atual = filepath.Dir(atual) {
		if _, err := os.Stat(atual); err == nil || atual == filepath.Dir(atual) {
			break
		}
		faltando = append(faltando, atual)
	}
	for i := len(faltando) - 1; i >= 0; i-- {
		if err := os.Mkdir(faltando[i], 0o755); err != nil {
			return err
		}
		if err := os.Chown(faltando[i], s.Conta.UID, s.Conta.GID); err != nil {
			return err
		}
	}
	return nil
}

// Escrever escreve (ou anexa) um arquivo.
func (s *Sistema) Escrever(caminho, conteudo string, o Escrita) error {
	acao := "escrever"
	if o.Anexar {
		acao = "anexar em"
	}
	if s.mostrar(acao, caminho) {
		return nil
	}
	if o.DoUsuario {
		if err := s.CriarDirUsuario(filepath.Dir(caminho)); err != nil {
			return err
		}
	}
	if o.Modo == 0 {
		o.Modo = 0o644
	}
	_, err := os.Stat(caminho)
	novo := errors.Is(err, fs.ErrNotExist)

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if o.Anexar {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(caminho, flags, o.Modo)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(conteudo); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if novo {
		// O umask pode ter cortado bits do modo pedido.
		if err := os.Chmod(caminho, o.Modo); err != nil {
			return err
		}
	}
	if o.DoUsuario {
		return os.Chown(caminho, s.Conta.UID, s.Conta.GID)
	}
	return nil
}

// InstalarConteudo equivale ao `install -o USER -g GRUPO -m MODO`, a partir do
// conteúdo já lido; origem serve só para o --dry-run.
func (s *Sistema) InstalarConteudo(origem string, conteudo []byte, destino string, modo os.FileMode) error {
	if s.mostrar(fmt.Sprintf("install -m %o %s ->", modo, origem), destino) {
		return nil
	}
	if err := s.CriarDirUsuario(filepath.Dir(destino)); err != nil {
		return err
	}
	if err := os.WriteFile(destino, conteudo, modo); err != nil {
		return err
	}
	if err := os.Chown(destino, s.Conta.UID, s.Conta.GID); err != nil {
		return err
	}
	return os.Chmod(destino, modo)
}

// CopiarPreservando equivale ao `cp -p`.
func (s *Sistema) CopiarPreservando(origem, destino string) error {
	if s.mostrar(fmt.Sprintf("cp -p %s ->", origem), destino) {
		return nil
	}
	info, err := os.Stat(origem)
	if err != nil {
		return err
	}
	conteudo, err := os.ReadFile(origem)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destino, conteudo, info.Mode().Perm()); err != nil {
		return err
	}
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := os.Chown(destino, int(st.Uid), int(st.Gid)); err != nil {
			return err
		}
	}
	if err := os.Chmod(destino, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Chtimes(destino, info.ModTime(), info.ModTime())
}

// Remover equivale ao rm -rf.
func (s *Sistema) Remover(caminho string) error {
	if s.mostrar("rm -rf", caminho) {
		return nil
	}
	return os.RemoveAll(caminho)
}

// Link equivale ao ln -sf.
func (s *Sistema) Link(alvo, link string) error {
	if s.mostrar(fmt.Sprintf("ln -sf %s ->", alvo), link) {
		return nil
	}
	if err := os.Remove(link); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return os.Symlink(alvo, link)
}
