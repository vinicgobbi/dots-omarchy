# post_omarchy

Pós-instalação exclusiva do [Omarchy](https://omarchy.org/) (Arch + Hyprland):
pacotes, apps, integrações nativas do Omarchy e os dots (Hyprland e
omarchy-shell). Para outras distros (Fedora, RHEL, Debian/Ubuntu, Arch puro),
use o projeto `post_install`.

## Uso

```bash
sudo ./setup.sh
```

Abre uma interface no terminal (TUI) que guia por etapas:

1. **Usuário** — escolhe a conta-alvo numa lista (ou digita outra, inclusive de
   domínio AD/LDAP);
2. **Módulos** — checklist com todos marcados; ao lado, a descrição, de quem o
   módulo depende e para quem ele é necessário. Marcar um módulo marca as
   dependências junto; desmarcar uma dependência ainda em uso é bloqueado com
   o motivo;
3. **Revisão** — usuário, modo (real ou dry-run), log e a lista final; os
   módulos marcados com ⌨ podem pedir senha ou confirmação;
4. **Execução** — barra de progresso, status de cada módulo e a saída dos
   comandos ao vivo (com rolagem). Passos que podem pedir a senha do sudo
   (`omarchy-update`, Chrome, bootstrap dos Dotfiles) recebem o terminal
   direto e a TUI volta sozinha ao fim. Plugins e temas de terceiros pedem
   confirmação numa janela com os riscos. `ctrl+c` duas vezes aborta;
5. **Resumo** — o que rodou, quanto tempo levou, avisos e onde estão os logs.

Recusa rodar em qualquer sistema que não tenha `ID=omarchy`. Cada execução
grava em `logs/` um log limpo (mensagens) e um `.raw.log` com a saída
completa de todos os comandos.

É escrito em Go ([Bubble Tea](https://github.com/charmbracelet/bubbletea) +
Lip Gloss). O `setup.sh` compila o binário `post-omarchy` na primeira vez (e
sempre que o código muda), instalando o pacote `go` via pacman se faltar.
Opções:

```bash
sudo ./setup.sh --dry-run                       # mostra o que seria feito, sem alterar nada (não exige root)
sudo ./setup.sh --modulos dots_omarchy,webapps  # já vem com esses (e as dependências) marcados, direto na revisão
sudo ./setup.sh --usuario vinicius              # pula a escolha do usuário
sudo ./setup.sh --sem-tui                       # saída em texto corrido, sem a interface
./setup.sh --listar                             # lista os módulos e seus ids
```

Sem terminal interativo (ex.: `ssh` sem tty, saída redirecionada), roda em
modo texto: todos os módulos ou os de `--modulos`, e plugins/temas de
terceiros **não** são instalados, por não haver como confirmar.

## Releases

Cada push na `main` passa pelo workflow `.github/workflows/release.yml`:

1. o [Commitizen](https://commitizen-tools.github.io/commitizen/) lê os commits
   desde a última tag e decide a versão (`feat` sobe o minor, `fix`/`refactor`
   o patch; enquanto for `0.x`, quebra de compatibilidade também sobe só o
   minor). Se só houver `docs`, `chore`, `ci` etc., não há release;
2. ele atualiza o `CHANGELOG.md`, faz o commit `bump: …` e cria a tag `vX.Y.Z`;
3. o binário é compilado nessa tag (estático, linux/amd64) e empacotado;
4. o release `vX.Y.Z` é publicado com as notas do changelog e dois arquivos:
   `post-omarchy-linux-amd64` (só o binário) e
   `post-omarchy-vX.Y.Z-linux-amd64.zip` (binário + `setup.sh`, `data/`,
   `dots/`, `assets/` e `OVPN/`, pronto para rodar sem Go);
5. o artefato intermediário da action é apagado; fica só o release.

Para usar o release numa máquina nova:

```bash
unzip post-omarchy-vX.Y.Z-linux-amd64.zip && cd post-omarchy-vX.Y.Z
sudo ./setup.sh
```

Os commits seguem o padrão do Commitizen (`cz commit` ajuda a montar a
mensagem; `cz check` valida). A configuração fica em `.cz.toml`. Para instalar:
`uv tool install commitizen` (ou `pipx install commitizen`).

## Estrutura

- `main.go` — flags, checa root e `ID=omarchy`, carrega `data/` e abre a TUI
  ou o modo texto; `setup.sh` compila e chama o binário
- `internal/tui/` — a interface: `modelo.go` (estado, teclas e a goroutine que
  roda os módulos), `telas.go` (desenho de cada etapa) e `ponte.go` (liga
  mensagens, perguntas e saída dos comandos dos módulos à tela)
- `internal/ui/` — a interface que os módulos usam (`Info`, `Aviso`,
  `Confirmar`…) e a implementação em texto corrido
- `internal/sistema/` — comandos como root ou como o usuário-alvo,
  `pacman`/`yay`/`flatpak`, escrita de arquivos e o `--dry-run`
- `internal/logs/` — logs em `logs/`; `internal/dados/` — lê `data/`
- `internal/modulos/` — um arquivo por etapa (`m01_atualizacao.go`…), e a
  ordem de execução em `modulo.go` (`Todos`). Cada um define um `*Modulo` com
  id, título, descrição, função, dependências e se é interativo
- `data/` — tudo que é lista fica aqui, em JSON:
  - `flatpaks.json` — apps (`apps`) e launchers de jogos (`jogos`)
  - `plugins.json` — plugins da barra (`omarchy plugin add`)
  - `themes.json` — temas extras; o último da lista fica ativo
  - `webapps.json` — nome, URL e ícone (URL ou caminho relativo ao projeto)
  - `packages.json` — pacotes de cada módulo (`pacman`, `aur`, `cargo`…), pelo id do módulo
  - `dots.json` — o que copiar de `dots/` para o home (origem, destino, modo e regras especiais)
- `dots/` — configs versionadas, aplicadas por `internal/modulos/m17_dots_omarchy.go`
  - `dots/hypr/` — `bindings.lua`, `input.lua`, `looknfeel.lua`, `monitors.lua`,
    `autostart.lua` (sobe o `fullscreen-dnd`) e `hyprland.lua` (jogos Proton em
    tela cheia sem apagar a tela) — só o que foi customizado; o resto fica no
    default do Omarchy
  - `dots/omarchy/` — `shell.json` (layout da barra), `branding/screensaver.txt`
    (screensaver personalizado) e
    `plugins/README.md` (lista dos plugins com link de origem)
  - `dots/solaar/` — `config.yaml` (desvio do botão de gesto e da thumb wheel do
    MX Master 3S) e `rules.yaml` (gestos → mídia/bloqueio, thumb wheel → volume
    via `solaar-volume`)
  - `dots/bin/` — scripts instalados em `~/.local/bin`: `fullscreen-dnd` (liga o
    "Silence Notifications" enquanto houver janela em tela cheia) e
    `solaar-volume` (volume da thumb wheel sem corrida entre eventos)
  - `dots/claude/` — `CLAUDE.md` global do Claude Code (não versionado, veja
    `dots/claude/README.md`)
- `OVPN/` — perfis `.ovpn` a importar (não versionados, veja `OVPN/README.md`)

## Módulos

| # | Módulo | O que faz |
|---|---|---|
| 01 | atualizacao | `omarchy-update -y` (snapshot, keyring, migrations, upgrade) |
| 02 | repositorios | Chaotic-AUR + sudo temporário do pacman para o `yay` (mirror do Omarchy é mantido) |
| 03 | pacotes_base | Docker, PHP/composer, unixODBC, VSCode (pacote do `[omarchy]`), MS SQL tools |
| 04 | solaar | Solaar + regras UDEV |
| 05 | flatpaks | Spotify nativo + lista de Flatpaks |
| 06 | launchers_jogos | Steam/Heroic nativos + drivers lib32; ProtonPlus/PrismLauncher via Flatpak |
| 07 | tailscale | Tailscale + widget na barra, webapp do admin console e Taildrop |
| 08 | chrome_gcm | `omarchy-install-browser chrome` + Git Credential Manager |
| 09 | bitwarden | Bitwarden desktop (AUR) |
| 10 | php_extensoes | `sqlsrv`/`pdo_sqlsrv` |
| 11 | ambiente_usuario | Zsh, Oh My Zsh (via repo Dotfiles), Node via mise, atalhos |
| 12 | rust_tools | rustup, eza, topgrade |
| 13 | claude_code | Pula se o Omarchy já instalou via mise |
| 14 | ovpn | Importa os `.ovpn` no NetworkManager |
| 15 | virt_manager | QEMU/KVM + libvirt |
| 16 | vscode_nautilus | "Abrir com o VSCode", "Copiar caminho" e "Abrir no Terminal" (terminal padrão do Omarchy) no Nautilus |
| 17 | dots_omarchy | Aplica `dots/` em `~/.config`, os scripts em `~/.local/bin` e o `CLAUDE.md` em `~/.claude`, se houver (com backup `.bak-post-omarchy`) |
| 18 | webapps | Apaga os webapps de fábrica e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, GitHub, Claude, GLPI e o admin console do Tailscale |
| 19 | limpeza | Órfãos, cache e remove o sudo temporário do `yay` |
| 20 | plugins_omarchy | Mostra os riscos e **pergunta** antes de instalar os plugins de terceiros da barra (`omarchy plugin add`) |
| 21 | temas_omarchy | **Pergunta** antes de instalar os temas extras (`omarchy-theme-install`); o último da lista, `lunar-quest`, fica ativo |

## Dots

`17_dots_omarchy` copia os arquivos de `dots/` para o home seguindo o
`data/dots.json`. Se o arquivo de destino já existir e for diferente, o original
é guardado uma vez como `<arquivo>.bak-post-omarchy`.

Cada entrada do `dots.json` tem `origem` (relativa a `dots/`, aceita glob como
`hypr/*.lua`) e `destino` (relativo ao home; termina em `/` quando a origem é um
glob). Opcionais: `modo` (padrão `0644`), `trocar_home` (troca `/home/vinicius/`
pelo home do usuário-alvo) e `opcional` + `aviso_ausente` (pula com aviso se o
arquivo não existir). Para versionar um dot novo, coloque o arquivo em `dots/`
e adicione a entrada.

Para atualizar os dots a partir da máquina, copie de volta os arquivos de
`~/.config/hypr`, `~/.config/omarchy` e `~/.config/solaar` para `dots/` (e os
scripts de `~/.local/bin` para `dots/bin/`).
