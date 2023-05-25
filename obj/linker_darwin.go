package obj

import (
    `os`
    `os/exec`
    `strings`
)

// TargetOS represents the target OS for linking.
const TargetOS = MacOS

func link(dest string, file string) error {
    var err error
    var sdk []byte

    /* get the SDK path */
    if sdk, err = exec.Command("xcrun", "-sdk", "macosx", "--show-sdk-path").Output(); err != nil {
        return err
    }

    /* construct the linker command */
    cmd := exec.Command(
        findldcmd(),
        "-lSystem",
        "-syslibroot",
        strings.TrimSpace(string(sdk)),
        "-o",
        dest,
        file,
    )

    /* bind stdio */
    cmd.Stdin  = nil
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
