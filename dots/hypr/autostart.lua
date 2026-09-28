-- Extra autostart processes.
-- o.launch_on_start("my-service")
o.launch_on_start(os.getenv("HOME") .. "/.local/bin/fullscreen-dnd")
