// Package atualizacao confere se há uma release mais nova do post-omarchy no
// GitHub e, se o usuário quiser, substitui o binário (e, quando ele roda de
// dentro do pacote da release, também setup.sh, data/, dots/, assets/ e os
// READMEs) pela versão nova. O processo atual continua com a versão antiga: é
// preciso executar de novo depois de atualizar.
package atualizacao

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	repo    = "vinicgobbi/dots-omarchy"
	binario = "post-omarchy"
	// maxArquivo limita cada arquivo extraído, para um pacote adulterado não
	// lotar o disco.
	maxArquivo = 256 << 20
)

// Release é uma versão nova disponível.
type Release struct {
	Tag    string
	url    string // .zip da release
	sumURL string // .sha256 do .zip
	digest string // SHA-256 calculado pelo GitHub no upload (pode faltar)
}

func nomePacote(tag string) string { return binario + "-" + tag + "-linux-amd64.zip" }

// semDowngrade recusa redirecionamentos de https:// para http://.
func semDowngrade(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("redirecionamentos demais")
	}
	if req.URL.Scheme != "https" {
		return fmt.Errorf("redirecionamento inseguro para %s", req.URL)
	}
	return nil
}

var cliente = &http.Client{CheckRedirect: semDowngrade}

func get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", binario)
	resposta, err := cliente.Do(req)
	if err != nil {
		return nil, err
	}
	if resposta.StatusCode >= 400 {
		resposta.Body.Close()
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resposta.StatusCode)
	}
	return resposta, nil
}

// Verificar devolve a última release se ela for mais nova que atual, ou nil se
// não houver nada novo. Builds locais (versão "dev" ou fora do formato vX.Y.Z)
// não são conferidos.
func Verificar(ctx context.Context, atual string) (*Release, error) {
	if _, ok := lerVersao(atual); !ok {
		return nil, nil
	}
	resposta, err := get(ctx, "https://api.github.com/repos/"+repo+"/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resposta.Body.Close()
	var dados struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resposta.Body, 4<<20)).Decode(&dados); err != nil {
		return nil, err
	}
	if !MaisNova(dados.Tag, atual) {
		return nil, nil
	}
	r := &Release{Tag: dados.Tag}
	pacote := nomePacote(dados.Tag)
	for _, a := range dados.Assets {
		switch a.Name {
		case pacote:
			r.url = a.URL
			r.digest, _ = strings.CutPrefix(a.Digest, "sha256:")
		case pacote + ".sha256":
			r.sumURL = a.URL
		}
	}
	if r.url == "" || r.sumURL == "" {
		return nil, fmt.Errorf("a release %s não tem %s e o .sha256 dele", dados.Tag, pacote)
	}
	return r, nil
}

// lerVersao lê "vX.Y.Z" (o "v" é opcional).
func lerVersao(v string) ([3]int, bool) {
	var numeros [3]int
	partes := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(partes) != 3 {
		return numeros, false
	}
	for i, p := range partes {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return numeros, false
		}
		numeros[i] = n
	}
	return numeros, true
}

// MaisNova diz se a versão a é maior que b (ambas no formato vX.Y.Z).
func MaisNova(a, b string) bool {
	va, ok1 := lerVersao(a)
	vb, ok2 := lerVersao(b)
	if !ok1 || !ok2 {
		return false
	}
	for i := range va {
		if va[i] != vb[i] {
			return va[i] > vb[i]
		}
	}
	return false
}

// Aplicar baixa a release, confere o SHA-256 e substitui o executável exe. Se
// exe estiver dentro do pacote da release (data/ ao lado e sem o código-fonte),
// atualiza também o resto do pacote. Os .ovpn, os logs e o CLAUDE.md nunca são
// tocados, porque não vêm no pacote. Os arquivos novos ficam com o dono do
// executável atual (o programa roda como root, mas o pacote é do usuário).
func Aplicar(ctx context.Context, r *Release, exe string) error {
	base := filepath.Dir(exe)
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	uid, gid := -1, -1
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		uid, gid = int(st.Uid), int(st.Gid)
	}
	_, errCodigo := os.Stat(filepath.Join(base, "main.go"))
	completo := eDiretorio(filepath.Join(base, "data")) && errCodigo != nil

	// Tudo é preparado num diretório temporário na mesma pasta (mesmo sistema
	// de arquivos, para o rename final ser atômico) e só então movido para o
	// lugar, com o binário por último.
	preparo, err := os.MkdirTemp(base, ".post-omarchy-update-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(preparo)

	pacote := filepath.Join(preparo, nomePacote(r.Tag))
	if err := baixar(ctx, r.url, pacote); err != nil {
		return err
	}
	if err := conferir(ctx, r, pacote); err != nil {
		return err
	}
	arquivos, err := extrair(pacote, filepath.Join(preparo, "pacote"), binario+"-"+r.Tag+"/", completo)
	if err != nil {
		return err
	}
	if _, ok := arquivos[binario]; !ok {
		return errors.New("o pacote da release não contém o binário")
	}

	for rel, origem := range arquivos {
		if rel == binario {
			continue
		}
		destino := filepath.Join(base, filepath.FromSlash(rel))
		if err := criarDirs(base, filepath.Dir(destino), uid, gid); err != nil {
			return err
		}
		modo := os.FileMode(0o644)
		if rel == "setup.sh" {
			modo = 0o755
		}
		if err := mover(origem, destino, modo, uid, gid); err != nil {
			return err
		}
	}
	return mover(arquivos[binario], exe, 0o755, uid, gid)
}

func baixar(ctx context.Context, url, destino string) error {
	ctx, cancelar := context.WithTimeout(ctx, 10*time.Minute)
	defer cancelar()
	resposta, err := get(ctx, url)
	if err != nil {
		return err
	}
	defer resposta.Body.Close()
	f, err := os.OpenFile(destino, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, io.LimitReader(resposta.Body, 1<<30)); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// conferir compara o SHA-256 do pacote com o .sha256 publicado na release e com
// o digest que o GitHub calculou no upload.
func conferir(ctx context.Context, r *Release, pacote string) error {
	ctx, cancelar := context.WithTimeout(ctx, 30*time.Second)
	defer cancelar()
	resposta, err := get(ctx, r.sumURL)
	if err != nil {
		return err
	}
	corpo, err := io.ReadAll(io.LimitReader(resposta.Body, 4096))
	resposta.Body.Close()
	if err != nil {
		return err
	}
	campos := strings.Fields(string(corpo))
	if len(campos) == 0 {
		return errors.New("arquivo .sha256 da release vazio")
	}

	f, err := os.Open(pacote)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	obtido := hex.EncodeToString(h.Sum(nil))
	for _, esperado := range []string{campos[0], r.digest} {
		if esperado != "" && !strings.EqualFold(obtido, esperado) {
			return fmt.Errorf("SHA-256 do pacote (%s) não confere com o da release (%s)", obtido, esperado)
		}
	}
	return nil
}

// permitido diz se um caminho do pacote (relativo à pasta raiz dele) deve ser
// atualizado.
func permitido(rel string, completo bool) bool {
	if rel == binario {
		return true
	}
	if !completo {
		return false
	}
	return rel == "README.md" || rel == "setup.sh" || rel == "OVPN/README.md" ||
		strings.HasPrefix(rel, "data/") || strings.HasPrefix(rel, "dots/") || strings.HasPrefix(rel, "assets/")
}

// extrair descompacta os arquivos permitidos do pacote em pasta e devolve
// caminho relativo -> arquivo extraído. Só arquivos regulares; links, caminhos
// absolutos ou com ".." são ignorados.
func extrair(pacote, pasta, raiz string, completo bool) (map[string]string, error) {
	z, err := zip.OpenReader(pacote)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	arquivos := map[string]string{}
	for _, f := range z.File {
		if !f.Mode().IsRegular() {
			continue
		}
		rel, ok := strings.CutPrefix(path.Clean(f.Name), raiz)
		if !ok || rel == "" || strings.HasPrefix(rel, "../") || path.IsAbs(rel) || !permitido(rel, completo) {
			continue
		}
		if f.UncompressedSize64 > maxArquivo {
			return nil, fmt.Errorf("%s no pacote é grande demais (%d bytes)", rel, f.UncompressedSize64)
		}
		destino := filepath.Join(pasta, filepath.FromSlash(rel))
		if err := extrairArquivo(f, destino); err != nil {
			return nil, err
		}
		arquivos[rel] = destino
	}
	return arquivos, nil
}

func extrairArquivo(f *zip.File, destino string) error {
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return err
	}
	leitor, err := f.Open()
	if err != nil {
		return err
	}
	defer leitor.Close()
	saida, err := os.OpenFile(destino, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(saida, io.LimitReader(leitor, maxArquivo)); err != nil {
		saida.Close()
		return err
	}
	return saida.Close()
}

// criarDirs cria as pastas que faltarem de base até pasta, com o dono uid/gid
// (as que já existem ficam como estão).
func criarDirs(base, pasta string, uid, gid int) error {
	rel, err := filepath.Rel(base, pasta)
	if err != nil || rel == "." {
		return err
	}
	atual := base
	for _, parte := range strings.Split(rel, string(filepath.Separator)) {
		atual = filepath.Join(atual, parte)
		if err := os.Mkdir(atual, 0o755); errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return err
		}
		if uid >= 0 {
			if err := os.Chown(atual, uid, gid); err != nil {
				return err
			}
		}
	}
	return nil
}

// mover põe origem no lugar de destino (rename atômico), com a permissão e o
// dono uid/gid dados.
func mover(origem, destino string, modo os.FileMode, uid, gid int) error {
	if err := os.Chmod(origem, modo); err != nil {
		return err
	}
	if uid >= 0 {
		if err := os.Chown(origem, uid, gid); err != nil {
			return err
		}
	}
	return os.Rename(origem, destino)
}

func eDiretorio(caminho string) bool {
	info, err := os.Stat(caminho)
	return err == nil && info.IsDir()
}
