package sistema

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// semDowngrade recusa redirecionamentos de https:// para http://, que tirariam
// a proteção do TLS no meio do caminho.
func semDowngrade(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("redirecionamentos demais")
	}
	if via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("redirecionamento inseguro de %s para %s", via[0].URL, req.URL)
	}
	return nil
}

var (
	clienteDownload = &http.Client{Timeout: 10 * time.Minute, CheckRedirect: semDowngrade}
	// clienteGet não tem timeout próprio: cada Get define o seu pelo contexto.
	clienteGet = &http.Client{CheckRedirect: semDowngrade}
)

// maxGet limita o corpo lido por Get (APIs e .sha256): nada legítimo aqui chega
// perto disso.
const maxGet = 32 << 20

// Get faz uma requisição somente-leitura com timeout e devolve o corpo. Falha
// em status HTTP de erro (equivale ao curl -f).
func (s *Sistema) Get(url string, prazo time.Duration) ([]byte, error) {
	ctx, cancelar := context.WithTimeout(s.ctx, prazo)
	defer cancelar()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "post-omarchy")
	resposta, err := clienteGet.Do(req)
	if err != nil {
		return nil, err
	}
	defer resposta.Body.Close()
	if resposta.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: HTTP %d", url, resposta.StatusCode)
	}
	corpo, err := io.ReadAll(io.LimitReader(resposta.Body, maxGet+1))
	if err == nil && len(corpo) > maxGet {
		return nil, fmt.Errorf("GET %s: resposta maior que %d MiB", url, maxGet>>20)
	}
	return corpo, err
}

// Baixar salva uma URL em destino, com algumas tentativas para tolerar
// instabilidade de rede.
func (s *Sistema) Baixar(url, destino string) error {
	if s.mostrar("baixar "+url+" ->", destino) {
		return nil
	}
	const tentativas = 3
	var err error
	for i := 1; i <= tentativas; i++ {
		if err = s.baixarUmaVez(url, destino); err == nil {
			return nil
		}
		if s.ctx.Err() != nil {
			return s.ctx.Err()
		}
		s.UI.Info(fmt.Sprintf("Falha ao baixar %s (tentativa %d/%d): %v", url, i, tentativas, err))
		time.Sleep(time.Duration(i) * time.Second)
	}
	return fmt.Errorf("não foi possível baixar %s após %d tentativas: %w", url, tentativas, err)
}

func (s *Sistema) baixarUmaVez(url, destino string) error {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "post-omarchy")
	resposta, err := clienteDownload.Do(req)
	if err != nil {
		return err
	}
	defer resposta.Body.Close()
	if resposta.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resposta.StatusCode)
	}
	parcial := destino + ".part"
	f, err := os.Create(parcial)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resposta.Body); err != nil {
		f.Close()
		os.Remove(parcial)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// O usuário-alvo também precisa ler (ex.: deploy-dev).
	if err := os.Chmod(parcial, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(s.Log, "download %s -> %s\n", url, destino)
	return os.Rename(parcial, destino)
}

// BaixarVerificado é o Baixar conferindo o SHA-256 (hex) do arquivo baixado.
func (s *Sistema) BaixarVerificado(url, destino, sha string) error {
	if err := s.Baixar(url, destino); err != nil || s.DryRun {
		return err
	}
	if err := VerificarSHA256(destino, sha); err != nil {
		os.Remove(destino)
		return err
	}
	return nil
}

// VerificarSHA256 compara o SHA-256 de um arquivo com o esperado (hex).
func VerificarSHA256(caminho, esperado string) error {
	f, err := os.Open(caminho)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if obtido := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(obtido, esperado) {
		return fmt.Errorf("SHA-256 de %s (%s) não confere com o esperado (%s)", filepath.Base(caminho), obtido, esperado)
	}
	return nil
}
