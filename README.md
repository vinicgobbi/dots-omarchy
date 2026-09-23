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

## Estrutura

- `setup.sh` — orquestrador (menu, dependências, execução, log)
- `config.sh` — listas de Flatpaks, launchers de jogos e plugins do shell
- `lib/` — `ui.sh` (menu/log/mensagens), `utils.sh` (registro de módulos,
  `yay`, downloads) e `os_detect.sh` (checa `ID=omarchy`)
- `modules/` — um arquivo por etapa, na ordem numérica de execução
- `dots/` — configs versionadas, aplicadas por `modules/17_dots_omarchy.sh`
  - `dots/hypr/` — `bindings.lua`, `input.lua`, `looknfeel.lua`, `monitors.lua` (só o que foi
    customizado; o resto fica no default do Omarchy)
  - `dots/omarchy/` — `shell.json` (layout da barra), `branding/screensaver.txt`
    (screensaver personalizado) e
    `plugins/README.md` (lista dos plugins com link de origem)
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
| 16 | vscode_nautilus | "Abrir com o VSCode" e "Copiar caminho" no Nautilus |
| 17 | dots_omarchy | Aplica `dots/` em `~/.config` e o `CLAUDE.md` em `~/.claude`, se houver (com backup `.bak-post-omarchy`) |
| 18 | webapps | Apaga os webapps de fábrica e cria YouTube, WhatsApp, Gmail, Netflix, Twitch, GitHub, GLPI e o admin console do Tailscale |
| 19 | limpeza | Órfãos, cache e remove o sudo temporário do `yay` |
| 20 | plugins_omarchy | Mostra os riscos e **pergunta** antes de instalar os plugins de terceiros da barra (`omarchy plugin add`) |

## Dots

`17_dots_omarchy` copia os arquivos de `dots/` para `~/.config`. Se o arquivo
de destino já existir e for diferente, o original é guardado uma vez como
`<arquivo>.bak-post-omarchy`.

Para atualizar os dots a partir da máquina, copie de volta os arquivos de
`~/.config/hypr` e `~/.config/omarchy` para `dots/`.
