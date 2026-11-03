# post_omarchy

Pós-instalação exclusiva do [Omarchy](https://omarchy.org/) (Arch + Hyprland):
pacotes, apps, integrações nativas do Omarchy e os dots (Hyprland e
omarchy-shell). Para o Ubuntu com GNOME, use o projeto `post-install`.

## Uso

```bash
sudo ./setup.sh
```

Abre uma interface no terminal (TUI) que guia por etapas:

1. **Nova versão** — só se houver uma release mais nova: atualizar agora ou
   ignorar (veja [Atualização](#atualização));
2. **Usuário** — escolhe a conta-alvo numa lista (ou digita outra, inclusive de
   domínio AD/LDAP). Quem executou o programa (`SUDO_USER`, `PKEXEC_UID` ou
   `logname`) já vem selecionado;
3. **Tema** — qual tema extra fica ativo, ou não instalar nenhum;
4. **Módulos** — checklist agrupado, com todos marcados; ao lado, a descrição,
   de quem o módulo depende e para quem ele é necessário. Marcar um módulo marca
   as dependências junto; desmarcar uma dependência desmarca quem depende dela,
   com um aviso na tela;
5. **Revisão** — usuário, modo (real ou dry-run), log, tema e a lista final; os
   módulos marcados com `?` perguntam antes de instalar código de terceiros;
6. **Execução** — barra de progresso, status de cada módulo (✓ ok, ! com avisos,
   ✗ falhou, ⤼ pulado, ⊘ cancelado) e a saída dos comandos ao vivo (com
   rolagem), tudo dentro da interface. Plugins e temas de terceiros pedem
   confirmação numa janela com os riscos. `ctrl+c` duas vezes aborta (`esc`
   desiste);
7. **Resumo** — o que rodou, quanto tempo levou, erros, avisos e onde estão os
   logs. O mesmo resumo fica impresso no terminal depois que a interface fecha.

Se um módulo falha, os seguintes continuam rodando, exceto os que dependem dele,
que são pulados. Use `--fail-fast` para parar tudo no primeiro erro.

Recusa rodar em qualquer sistema que não tenha `ID=omarchy`.

**Sudo sem senha temporário:** os scripts do Omarchy (`omarchy-update`,
`omarchy-install-browser`…) e o `yay` chamam `sudo` sozinhos e, rodando como o
usuário-alvo via `su -c`, não têm terminal para pedir a senha. Por isso, só
durante a execução, o setup cria `/etc/sudoers.d/99-post-omarchy` com
`NOPASSWD: ALL` e `verifypw=any` (para o `sudo -v` também não pedir senha) para o
usuário-alvo (que já autenticou no `sudo ./setup.sh`). A regra é removida ao fim, em falha, em abort, em SIGTERM/SIGHUP e, se uma
execução anterior tiver sido morta com SIGKILL, no início da próxima. Enquanto
ela existe, qualquer script rodando como esse usuário tem sudo sem senha,
inclusive os de terceiros (Dotfiles, rustup, plugins). Cada execução
grava em `logs/` um log limpo (mensagens) e um `.raw.log` com a saída
completa de todos os comandos, legíveis só pelo usuário-alvo (permissão `0600`).

É escrito em Go ([Bubble Tea](https://github.com/charmbracelet/bubbletea) +
Lip Gloss). O `setup.sh` compila o binário `post-omarchy` na primeira vez (e
sempre que o código muda), instalando o pacote `go` via pacman se faltar e
baixando a versão do `go.mod` se a do sistema for mais antiga.
Opções:

```bash
sudo ./setup.sh --dry-run                       # mostra o que seria feito, sem alterar nada (não exige root)
sudo ./setup.sh --modulos dots_omarchy,webapps  # já vem com esses (e as dependências) marcados, direto na revisão
sudo ./setup.sh --pular docker,ruby             # desmarca esses (e os que dependem deles)
sudo ./setup.sh --usuario vinicius              # pula a escolha do usuário
sudo ./setup.sh --tema lunar-quest              # pula a escolha do tema (ou "none" para não instalar temas)
sudo ./setup.sh --sem-tui                       # saída em texto corrido, sem a interface
sudo ./setup.sh --sem-tui --yes                 # no modo texto, começa sem pedir confirmação
sudo ./setup.sh --fail-fast                     # para tudo no primeiro módulo que falhar
sudo ./setup.sh --sem-atualizacao               # não verifica se há versão nova
./setup.sh --listar                             # lista os módulos e seus ids
```

Sem terminal interativo (ex.: `ssh` sem tty, saída redirecionada), roda em
modo texto: todos os módulos ou os de `--modulos`, e plugins/temas de
terceiros **não** são instalados, por não haver como confirmar. O `--yes` só
pula a confirmação de início: plugins e temas continuam perguntando.

## Atualização

Ao iniciar, um binário de release consulta a última release no GitHub (com prazo
de 5 s; sem rede, só avisa e segue). Se houver uma versão mais nova:

- **TUI** — a primeira tela oferece *Atualizar agora* ou *Ignorar e continuar*;
- **modo texto** — pergunta `Atualizar agora? [S/n]`;
- **`--dry-run`, `--yes` ou sem terminal** — só avisa, sem atualizar.

Atualizar baixa o `.zip` da release, confere o SHA-256 (pelo `.sha256` publicado
e pelo digest que o GitHub calcula) e substitui o binário. Dentro do pacote da
release, atualiza também `setup.sh`, `data/`, `dots/`, `assets/`, `README.md` e
`OVPN/README.md` (edições suas nesses arquivos são sobrescritas; um arquivo que
saiu da release fica onde está). Os `.ovpn`, os `logs/` e o `CLAUDE.md` nunca são
tocados, e os arquivos ficam com o dono do binário atual. A execução em andamento
continua sendo a versão antiga: ao final, o programa mostra o comando para
executar de novo.

Builds locais (clone do repositório, versão `dev`) não verificam atualização.

## Releases

Cada push na `main` passa pelo workflow `.github/workflows/release.yml`:

1. roda `go vet`, `go test`, `staticcheck`, `govulncheck` e `shellcheck`;
2. o [Commitizen](https://commitizen-tools.github.io/commitizen/) lê os commits
   desde a última tag e decide a versão (`feat` sobe o minor, `fix`/`refactor`
   o patch; enquanto for `0.x`, quebra de compatibilidade também sobe só o
   minor). Se só houver `docs`, `chore`, `ci` etc., não há release;
3. ele atualiza o `CHANGELOG.md`, faz o commit `bump: …` e cria a tag `vX.Y.Z`;
4. o binário é compilado nessa tag (estático, linux/amd64) e empacotado, com
   `.sha256` e atestado de proveniência;
5. o release `vX.Y.Z` é publicado com as notas do changelog e os arquivos:
   `post-omarchy-linux-amd64` (só o binário),
   `post-omarchy-vX.Y.Z-linux-amd64.zip` (binário + `setup.sh`, `data/`,
   `dots/`, `assets/` e `OVPN/`, pronto para rodar sem Go) e o `.sha256` de cada um;
6. o artefato intermediário da action é apagado; fica só o release.

Para usar o release numa máquina nova, conferindo o download:

```bash
sha256sum -c post-omarchy-vX.Y.Z-linux-amd64.zip.sha256
gh attestation verify post-omarchy-vX.Y.Z-linux-amd64.zip --repo vinicgobbi/dots-omarchy
unzip post-omarchy-vX.Y.Z-linux-amd64.zip && cd post-omarchy-vX.Y.Z
sudo ./setup.sh
```

Os commits seguem o padrão do Commitizen (`cz commit` ajuda a montar a
mensagem; `cz check` valida). A configuração fica em `.cz.toml`. Para instalar:
`uv tool install commitizen` (ou `pipx install commitizen`).

### Atualização das versões fixadas

Tudo o que é fixado tem quem avise quando sai versão nova, com um PR na `main`:

- **Actions e módulos Go** — o Dependabot (`.github/dependabot.yml`), toda semana.
- **Commits dos repositórios pessoais e regra UDEV do Solaar** — o workflow
  `.github/workflows/pins.yml`, todo dia (ou à mão em Actions → Pins → Run
  workflow). Ele roda `scripts/atualizar-pins.sh` e abre o PR com links de
  comparação. Precisa de *Settings → Actions → General → Allow GitHub Actions to
  create and approve pull requests*.

## Segurança

- **Versões fixadas.** Os repositórios pessoais executados (Dotfiles e extensões
  do Nautilus) ficam fixados num commit (`internal/modulos/pins.go`), e a regra
  UDEV do Solaar vem embutida no binário, só gravada se o pacote não a trouxer.
- **Downloads conferidos.** Assets de release do GitHub (GCM, deploy-dev) são
  conferidos pelo SHA-256 que o GitHub publica. O cliente HTTP recusa
  redirecionamento de `https` para `http` e tem prazo. O Chaotic-AUR é conferido
  pelo fingerprint completo da chave.
- **Temporários exclusivos.** Downloads e clones usam diretórios criados por
  `mktemp`, nunca nomes fixos em `/tmp`; a limpeza final não mexe em `/tmp`.
- **O que não é verificado.** Plugins e temas de terceiros vêm da versão atual de
  cada repositório, depois de uma janela com os riscos. rustup e Claude Code são
  instalados pelo `curl | sh` oficial, como o usuário-alvo, enquanto a regra
  temporária de sudo está ativa.
- **Grupos com poder de root.** O usuário-alvo entra nos grupos `docker` e
  `libvirt`/`kvm`: quem está neles consegue virar root (é o normal dessas
  ferramentas numa workstation de desenvolvimento, mas vale saber).

## Estrutura

- `main.go` — flags, checa root e `ID=omarchy`, carrega `data/` e abre a TUI
  ou o modo texto; `setup.sh` compila e chama o binário
- `internal/tui/` — a interface: `modelo.go` (estado, teclas e a goroutine que
  roda os módulos), `telas.go` (desenho de cada etapa), `atualizacao.go` (tela de
  versão nova) e `ponte.go` (liga mensagens, perguntas e saída dos comandos dos
  módulos à tela, em lotes)
- `internal/atualizacao/` — verifica e aplica versões novas (release do GitHub)
- `internal/ui/` — a interface que os módulos usam (`Info`, `Aviso`,
  `Confirmar`…) e a implementação em texto corrido
- `internal/sistema/` — comandos como root ou como o usuário-alvo,
  `pacman`/`yay`/`flatpak`, escrita de arquivos, downloads conferidos e o
  `--dry-run`
- `internal/logs/` — logs em `logs/`; `internal/dados/` — lê `data/`
- `internal/modulos/` — um arquivo por módulo (`atualizacao.go`…), a ordem de
  execução em `modulo.go` (`Todos`), o motor que roda a fila em `execucao.go` e
  os commits fixados em `pins.go`. Cada um define um `*Modulo` com id, grupo,
  título, descrição, função, dependências e se é interativo
- `data/` — tudo que é lista fica aqui, em JSON:
  - `flatpaks.json` — apps (`apps`) e launchers de jogos (`jogos`)
  - `plugins.json` — plugins da barra (`omarchy plugin add`)
  - `themes.json` — temas extras; o último da lista é o padrão da tela de tema
  - `webapps.json` — nome, URL e ícone (URL ou caminho relativo ao projeto)
  - `packages.json` — pacotes de cada módulo (`pacman`, `aur`, `cargo`…), pelo id do módulo
  - `dots.json` — o que copiar de `dots/` para o home (origem, destino, modo e regras especiais)
- `dots/` — configs versionadas, aplicadas por `internal/modulos/dots_omarchy.go`
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

Os módulos ficam em grupos, mostrados nessa ordem na TUI e no `--listar`, e rodam
de cima para baixo. Cada um é autocontido: instala o que usa, então pode ser
marcado ou desmarcado sozinho. A exceção são os que usam a AUR (`bitwarden` e
`php_sqlserver`), que dependem de `repositorios`.

| Grupo | ID | O que faz |
|---|---|---|
| Sistema | `atualizacao` | `omarchy-update -y` (snapshot, keyring, migrations, upgrade) |
| | `repositorios` | Chaotic-AUR (mirror do Omarchy é mantido) |
| Terminal | `ferramentas` | git, curl, jq, bat, gh, tea, nmap, pipx, ffmpeg, bashtop, FUSE 2 |
| | `shell` | Zsh como shell padrão e os dotfiles (Oh My Zsh, plugins, fontes) |
| Aplicativos | `firefox_thunderbird` | Firefox e Thunderbird |
| | `chrome` | `omarchy-install-browser chrome` |
| | `bitwarden` | Bitwarden desktop (AUR) |
| | `flatpaks` | Spotify nativo + lista de Flatpaks |
| | `launchers_jogos` | Steam/Heroic nativos + drivers lib32; ProtonPlus/PrismLauncher via Flatpak |
| | `solaar` | Solaar, regras UDEV e início com a sessão, só na bandeja |
| Desenvolvimento | `git_gcm` | Git Credential Manager (SHA-256 conferido) |
| | `docker` | Docker, Buildx e Compose; serviço ativo e usuário no grupo `docker` |
| | `vscode` | VSCode (pacote do `[omarchy]`) com o tema do Omarchy |
| | `go` | Go |
| | `node` | Node LTS via mise, com o mise ativo no zsh |
| | `php_sqlserver` | PHP, Composer, unixODBC, MS SQL tools e `sqlsrv`/`pdo_sqlsrv` |
| | `rust` | rustup, eza, topgrade, cargo-update |
| | `ruby` | Ruby e, como gems do usuário, Bundler e Jekyll |
| | `claude_code` | Pula se o Omarchy já instalou via mise |
| | `deploy_dev` | deploy-dev (SHA-256 conferido) |
| Rede e virtualização | `tailscale` | Tailscale + widget na barra, webapp do admin console e Taildrop |
| | `firewall` | ufw + gufw (não mexe no estado nem nas regras) |
| | `ovpn` | Importa os `.ovpn` no NetworkManager (só substitui conexões do tipo VPN) |
| | `virt_manager` | QEMU/KVM + libvirt |
| Omarchy e aparência | `nautilus` | "Abrir com o VSCode", "Copiar caminho", "Abrir no Terminal" e a pasta Projetos |
| | `dots_omarchy` | Aplica `dots/` em `~/.config`, os scripts em `~/.local/bin` e o `CLAUDE.md` em `~/.claude`, se houver (com backup `.bak-post-omarchy`) |
| | `webapps` | Apaga os webapps de fábrica e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, GitHub, Claude, GLPI e o admin console do Tailscale |
| | `plugins_omarchy` | Mostra os riscos e **pergunta** antes de instalar os plugins de terceiros da barra (`omarchy plugin add`) |
| | `temas_omarchy` | **Pergunta** antes de instalar os temas extras (`omarchy-theme-install`); o escolhido fica ativo |
| Finalização | `limpeza` | Órfãos e cache do pacman |

## Dots

O módulo `dots_omarchy` copia os arquivos de `dots/` para o home seguindo o
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
