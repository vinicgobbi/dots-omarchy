## v0.1.0 (2026-10-08)

### Feat

- sync dots with system (shell bar, Solaar, new plugins, Claude webapp, themes module)
- sync dots with system (shell plugins, Proton rules, helper scripts)
- install Solaar gesture and thumb wheel config from dots
- bind Super+L to lock screen and move layout toggle to Super+Shift+L
- add Open in Terminal extension to Nautilus module
- add dots/claude for the global CLAUDE.md (not versioned)
- install custom screensaver from dots
- move plugin install to a last, opt-in module with risk warning
- rounded window corners (rounding = 8) in Hyprland dots
- add Twitch, GitHub and GLPI webapps (GLPI with custom SVG icon)
- install nautilus-copypath extension alongside vscode_nautilus
- always create Tailscale admin webapp from the webapps module
- add webapps module (YouTube, WhatsApp, Gmail, Netflix)
- add Omarchy Hyprland and shell custom configs
- initial Omarchy post-install (modules and dots)

### Fix

- create Chrome policy dir as root so the Omarchy theme applies
- use --locked on cargo install for reproducible builds

### Refactor

- rewrite setup in Go with an interactive TUI
- rewrite setup in Python with lists in data/*.json
- drop fnm and use mise (Omarchy default) for Node
