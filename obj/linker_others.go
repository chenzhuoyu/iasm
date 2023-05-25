// +build !linux,!darwin

package obj

import (
    `errors`
)

// CurrentOS is the OS that iasm is built for.
const CurrentOS = Unsupported

func link(_ string, _ string) error {
    return errors.New("unsupported operating system")
}