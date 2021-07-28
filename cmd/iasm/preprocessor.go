package main

import (
    `bytes`
    `os`
    `os/exec`
)

const (
    defaultCpp = "/usr/bin/cpp"
)

func preprocess(name string, defs []string) (string, error) {
    var err error
    var cpp string
    var out bytes.Buffer

    /* find the CPP command */
    if cpp = os.Getenv("CPP"); cpp == "" {
        cpp = defaultCpp
    }

    /* command arguments */
    args := []string {
        "-CC",
        "-nostdinc",
    }

    /* build definations */
    for _, def := range defs {
        args = append(args, "-D" + def)
    }

    /* construct the preprocessor command */
    cmd := exec.Command(
        defaultCpp,
        append(args, name)...,
    )

    /* bind stdio */
    cmd.Stdin  = nil
    cmd.Stdout = &out
    cmd.Stderr = os.Stderr

    /* run the preprocessor */
    if err = cmd.Run(); err != nil {
        return "", err
    } else {
        return string(out.Bytes()), nil
    }
}
