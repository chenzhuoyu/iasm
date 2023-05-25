// +build !linux,!darwin

package obj

import (
    `errors`
)

// TargetOS represents the target OS for linking.
const TargetOS = Unsupported

func link(_ string, _ string) error {
    return errors.New("unsupported operating system")
}