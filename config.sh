#!/usr/bin/env bash
# ==========================================
# Configuração compartilhada
# ==========================================

# Apps Flatpak (Spotify e Steam/Heroic ficam de fora: têm pacote nativo,
# ver modules/05_flatpaks.sh e modules/06_launchers_jogos.sh).
# com.mattjakeman.ExtensionManager também fica de fora: gerencia extensões
# do GNOME Shell, que não existe no Omarchy (Hyprland).
FLATPAKS=(
  com.getpostman.Postman com.github.tchx84.Flatseal
  com.obsproject.Studio dev.vencord.Vesktop
  io.dbeaver.DBeaverCommunity io.ente.auth io.github.flattool.Ignition io.github.flattool.Warehouse
  io.missioncenter.MissionCenter md.obsidian.Obsidian
  org.filezillaproject.Filezilla org.gaphor.Gaphor org.gnome.Boxes
  org.libreoffice.LibreOffice org.onlyoffice.desktopeditors org.qbittorrent.qBittorrent
  org.remmina.Remmina org.telegram.desktop org.videolan.VLC me.iepure.devtoolbox
  com.rtosta.zapzap
)

# ProtonPlus e PrismLauncher não têm instalador dedicado no Omarchy.
FLATPAKS_JOGOS=(
  com.vysp3r.ProtonPlus org.prismlauncher.PrismLauncher
)

# Plugins do shell do Omarchy (instalados via "omarchy plugin clone <url>").
# A posição de cada um na barra vem de dots/omarchy/shell.json.
OMARCHY_PLUGINS=(
  https://github.com/BibekBhusal0/omarchy-better-media
  https://github.com/dstankovd/omarchy-openvpn
  https://github.com/MariusGhizdavet/omarchy-rss
  https://github.com/jitendradara12/omaconnect
  https://github.com/Wian47/omarchy-removable-drives
  https://github.com/vinicgobbi/omarchy-plugin-power
  https://github.com/vinicgobbi/omarchy-plugin-vpn
)
