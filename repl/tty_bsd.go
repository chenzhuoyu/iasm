// +build darwin dragonfly freebsd netbsd openbsd

package repl

import (
	`syscall`
)

const ioctlCode = syscall.TIOCGETA
