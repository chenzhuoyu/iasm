// +build linux aix zos
// +build !appengine

package repl

import (
	`syscall`
)

const ioctlCode = syscall.TCGETS
