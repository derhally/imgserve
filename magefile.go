//go:build mage
// +build mage

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	//"github.com/carolynvs/magex/pkg"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const (
	EXE = "imgserve"

	osLinux   = "linux"
	osWindows = "windows"

	dirBuild   = "bin"
	dirRelease = "release"
)

var extraFlags = "-tags;release"

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build.Local

var targets = map[string][]string{
	osLinux:   []string{"amd64", "arm64"},
	osWindows: []string{"amd64"},
}

var Aliases = map[string]interface{}{
	"build":   Build.All,
	"package": Package.All,
}

type Build mg.Namespace

func (Build) Local() error {
	mg.Deps(Clean)
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return buildfor(runtime.GOOS, runtime.GOARCH, EXE, extraFlags)
}

func (Build) Linux() error {
	mg.Deps(Clean)
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return buildfor(osLinux, "amd64", EXE, extraFlags)
}

func (Build) Windows() error {
	mg.Deps(Clean)
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return buildfor(osWindows, "amd64", EXE, extraFlags)
}

func (Build) All() {
	deps := make([]interface{}, 0)

	for targetOS, archList := range targets {
		for _, arch := range archList {
			deps = append(deps, mg.F(buildfor, targetOS, arch, EXE, extraFlags))
		}
	}

	mg.Deps(deps...)
}

type Package mg.Namespace

func (Package) All() {
	// Declares target dependencies that will be run in parallel
	deps := make([]interface{}, 0)

	for targetOS, archList := range targets {
		for _, arch := range archList {
			deps = append(deps, mg.F(release, targetOS, arch))
		}
	}

	mg.Deps(deps...)
}

func (Package) Linux() error {
	return release(osWindows, "amd64")
}

func (Package) Windows() error {
	return release(osWindows, "amd64")
}

func release(targetOS, arch string) error {
	// package is dependent on build for the same OS/arch
	mg.Deps(Build.All)

	buildTarget := executablePath(targetOS, arch, EXE)
	packageTarget := packagePath(targetOS, arch, EXE)

	// check if built binaries have been updated since the last package execution
	if updated, err := target.Path(packageTarget, buildTarget); !updated || err != nil {
		return err
	}

	if err := os.MkdirAll(dirRelease, 0755); err != nil {
		return err
	}

	fmt.Printf("Packaging %s/%s\n", targetOS, arch)

	if targetOS == osWindows {
		return sh.Run("zip", "-j", packageTarget, buildTarget)
	}

	return sh.RunV("tar", "-czf", packageTarget, "-C", filepath.Dir(buildTarget), filepath.Base(buildTarget))
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	sh.Rm(dirBuild)
	sh.Rm(dirRelease)
}

func buildfor(os, arch, executable string, args string) error {
	command := []string{
		"build",
		"-o", executablePath(os, arch, executable),
	}
	command = append(command, strings.Split(args, ";")...)
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
	if os == osWindows {
		extension = ".exe"
	}
	return fmt.Sprintf("%s/%s-%s/%s%s", dirBuild, os, arch, executable, extension)
}

func packagePath(os, arch, executable string) string {
	filename := fmt.Sprintf("%s-%s-%s", executable, os, arch)
	if os == osWindows {
		filename += ".zip"
	} else {
		filename += ".tgz"
	}

	return fmt.Sprintf("%s/%s", dirRelease, filename)
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
