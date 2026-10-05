# post_omarchy

Pós-instalação exclusiva do [Omarchy](https://omarchy.org/) (Arch + Hyprland):
pacotes, apps, integrações nativas do Omarchy e os dots (Hyprland e
omarchy-shell). Para outras distros (Fedora, RHEL, Debian/Ubuntu, Arch puro),
use o projeto `post_install`.

## Uso

```bash
sudo ./setup.sh
```

O script pede o usuário-alvo, mostra um menu com todos os módulos marcados,
resolve dependências entre eles e registra o log em `logs/`. Recusa rodar em
qualquer sistema que não tenha `ID=omarchy`.

É escrito em Python 3, só com a biblioteca padrão (o Omarchy já traz o
Python), e o `setup.sh` só chama o `setup.py`. Opções:

```bash
sudo ./setup.sh --dry-run                     # mostra o que seria feito, sem alterar nada
sudo ./setup.sh --modulos dots_omarchy,webapps  # roda só esses (e as dependências), sem menu
```

## Estrutura

- `setup.py` — orquestrador (checa `ID=omarchy`, menu, dependências, execução, log);
  `setup.sh` é só um atalho para ele
- `post_omarchy/` — `ui.py` (mensagens/menu), `log.py` (logs em `logs/`),
  `sistema.py` (comandos como root ou como o usuário-alvo, `pacman`/`yay`/`flatpak`,
  escrita de arquivos e o `--dry-run`), `modulos.py` (carga e dependências) e
  `dados.py` (lê `data/`)
- `modules/` — um arquivo `.py` por etapa, na ordem numérica de execução; cada um
  define `MODULO = Modulo(id, título, descrição, função, dependências)`
- `data/` — tudo que é lista fica aqui, em JSON:
  - `flatpaks.json` — apps (`apps`) e launchers de jogos (`jogos`)
  - `plugins.json` — plugins da barra (`omarchy plugin add`)
  - `themes.json` — temas extras; o último da lista fica ativo
  - `webapps.json` — nome, URL e ícone (URL ou caminho relativo ao projeto)
  - `packages.json` — pacotes de cada módulo (`pacman`, `aur`, `cargo`…), pelo id do módulo
  - `dots.json` — o que copiar de `dots/` para o home (origem, destino, modo e regras especiais)
- `dots/` — configs versionadas, aplicadas por `modules/17_dots_omarchy.py`
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
