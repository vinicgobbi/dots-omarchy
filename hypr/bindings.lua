-- Keep only your personal keybinding overrides here. Add new bindings or
-- unbind defaults before replacing them.

-- See current bindings and descriptions:
--   omarchy menu keybindings --print

-- To disable every Omarchy default binding, set this in
-- ~/.config/hypr/hyprland.lua before require("default.hypr.omarchy"), then add
-- only the bindings you want below:
--   omarchy_default_bindings = false

-- To disable all preinstalled app/webapp bindings, set:
--   omarchy_preinstalled_bindings = false

-- Add a new binding.
-- o.bind("SUPER + SHIFT + R", "SSH", "alacritty -e ssh your-server")

-- Change an existing binding by unbinding it first, then binding the key again.
-- This example changes SUPER+SPACE from the launcher to the Omarchy root menu.
-- hl.unbind("SUPER + SPACE")
-- o.bind("SUPER + SPACE", "Omarchy menu", "omarchy-menu toggle root")

-- Disable a default binding without replacing it.
-- hl.unbind("SUPER + SHIFT + B")

-- Logitech MX Keys examples:
-- o.bind("SUPER + SHIFT + S", nil, "omarchy-capture-screenshot")
-- o.bind("SUPER + H", nil, "voxtype record toggle")
-- o.bind("SUPER + PERIOD", nil, "omarchy-shell shell toggle omarchy.emojis")

-- Calendar -> Google Calendar (era app.hey.com)
hl.unbind("SUPER + SHIFT + C")
o.bind("SUPER + SHIFT + C", "Google Calendar", { webapp = "https://calendar.google.com/" })

-- Email -> Gmail (era app.hey.com)
hl.unbind("SUPER + SHIFT + E")
o.bind("SUPER + SHIFT + E", "Gmail", { webapp = "https://mail.google.com/mail/" })

-- New email -> Novo e-mail dentro do Gmail (era app.hey.com/messages/new)
hl.unbind("SUPER + SHIFT + ALT + E")
o.bind("SUPER + SHIFT + ALT + E", "New Gmail", { webapp = "https://mail.google.com/mail/?view=cm&fs=1" })

-- SUPER + SHIFT + S -> Screenshot (era Google Maps webapp), mesmo comando do PRINT
hl.unbind("SUPER + SHIFT + S")
o.bind("SUPER + SHIFT + S", "Screenshot", "omarchy-capture-screenshot")

-- SUPER + SHIFT + R -> Gravação de tela (mesmo comando do ALT+PRINT)
o.bind("SUPER + SHIFT + R", "Screenrecording", "omarchy-capture-screenrecording --stop-recording || omarchy-menu toggle trigger.capture.screenrecord")

-- Inverte SUPER + SPACE (Apps menu) com SUPER + ALT + SPACE (Omarchy menu)
hl.unbind("SUPER + SPACE")
hl.unbind("SUPER + ALT + SPACE")
o.bind("SUPER + SPACE", "Apps menu", "omarchy-menu toggle apps")
o.bind("SUPER + ALT + SPACE", "Omarchy menu", "omarchy-menu toggle")
