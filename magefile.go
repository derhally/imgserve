//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/magefile/mage/sh"
)

const (
	EXE = "picserve"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	return build(runtime.GOOS, runtime.GOARCH, EXE, []string{"-tags", "release"})
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin")
}

func build(os, arch, executable string, args []string) error {
	command := []string{
		"build",
		"-o", executablePath(os, arch, executable),
	}
	command = append(command, args...)
	command = append(command, ".")
	output, err := run("go", command, map[string]string{
		"GOOS":   os,
		"GOARCH": arch,
	})
	fmt.Print(output)
	return err
}

func executablePath(os, arch, executable string) string {
	extension := ""
	if os == "windows" {
		extension = ".exe"
	}
	return fmt.Sprintf("bin/%s-%s/%s%s", os, arch, executable, extension)
}

func run(program string, args []string, env map[string]string) (string, error) {
	// Make string representation of command
	fullArgs := append([]string{program}, args...)
	cmdStr := strings.Join(fullArgs, " ")

	// Make string representation of environment
	envStrBuf := new(bytes.Buffer)
	for key, value := range env {
		fmt.Fprintf(envStrBuf, "%s=\"%s\", ", key, value)
	}
	envStr := string(bytes.TrimRight(envStrBuf.Bytes(), ", "))

	// Show info
	fmt.Println("Running '" + cmdStr + "'" + " with env " + envStr)

	// Run
	return sh.OutputWith(env, program, args...)
}
