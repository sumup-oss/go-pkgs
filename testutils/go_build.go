package testutils

import (
	"bytes"
	"io/ioutil"
	stdOs "os"
	"os/exec"
	"runtime"

	log "github.com/sumup-oss/go-pkgs/logger"
	"github.com/sumup-oss/go-pkgs/os"
)

type Build struct {
	binaryPath string
	workDir    string
}

func NewBuild(binaryPath, workDir string) *Build {
	return &Build{
		binaryPath: binaryPath,
		workDir:    workDir,
	}
}

func (b *Build) cmd(args ...string) *exec.Cmd {
	//nolint:gosec
	cmd := exec.Command(b.binaryPath, args...)
	cmd.Dir = b.workDir

	// NOTE: Inherit environment of the host/container running the binary,
	// to make sure we're not isolating factors.
	cmd.Env = stdOs.Environ()

	return cmd
}

func (b *Build) Run(args ...string) (string, string, error) {
	cmdInstance := b.cmd(args...)

	var stdoutBuffer bytes.Buffer
	var stdErrBuffer bytes.Buffer

	// NOTE: Don't need stdin.
	cmdInstance.Stdin = nil
	cmdInstance.Stdout = &stdoutBuffer
	cmdInstance.Stderr = &stdErrBuffer

	err := cmdInstance.Run()
	return stdoutBuffer.String(), stdErrBuffer.String(), err
}

func GoBuild(binaryPattern, pkgPath string, osExecutor os.OsExecutor) string {
	tmpFile, err := ioutil.TempFile("", binaryPattern)
	if err != nil {
		log.Fatal(err)
	}

	tmpFilename := tmpFile.Name()

	err = tmpFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	// NOTE: On windows the temp file created in the previous step cannot be overwritten
	err = osExecutor.Remove(tmpFilename)
	if err != nil {
		log.Fatal(err)
	}

	if runtime.GOOS == "windows" {
		tmpFilename += ".exe"
	}

	cmd := exec.Command(
		"go",
		"build",
		"-v",
		"-o",
		tmpFilename,
		pkgPath,
	)
	cmd.Stderr = osExecutor.Stderr()
	// NOTE: Don't need stdin.
	cmd.Stdin = nil
	cmd.Stdout = osExecutor.Stderr()

	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to build executable: %s", err)
	}

	return tmpFilename
}
