# Memória global do Claude Code

Coloque aqui o seu `CLAUDE.md` global (o que fica em `~/.claude/CLAUDE.md`,
com as instruções que valem para todos os projetos):

```bash
cp ~/.claude/CLAUDE.md dots/claude/CLAUDE.md
```

O `modules/17_dots_omarchy.sh` instala esse arquivo em `~/.claude/CLAUDE.md`
durante o setup. Se já existir um diferente, o original é guardado uma vez
como `CLAUDE.md.bak-post-omarchy`. Sem o arquivo aqui, o módulo só pula
essa etapa.

O `CLAUDE.md` é pessoal e **não é versionado**: só este README fica no
repositório (veja `.gitignore`).
