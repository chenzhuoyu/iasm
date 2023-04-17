// +build darwin freebsd openbsd netbsd dragonfly
// +build !appengine

package repl

import (
	`syscall`
)

const ioctlCode = syscall.TIOCGETA
