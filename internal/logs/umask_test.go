package logs

import "syscall"

func unixUmask(mascara int) int { return syscall.Umask(mascara) }
