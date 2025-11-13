//go:build windows

package psmodule

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/coreos/go-semver/semver"
	"github.com/google/uuid"
)

type extractedFiles struct {
	installScript string
	packageFile   string
	cleanup       func()
}

var InstallLog = log.New(os.Stdout, "", log.LstdFlags)

func extractEmbeddedFiles() (*extractedFiles, error) {

	extractDir := uuid.NewString()

	if err := os.Mkdir(extractDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot make temp dir: %w", err)
	}

	packageToInstall, installScript, err := extractPackage(extractDir)

	if err != nil {
		return nil, err
	}

	return &extractedFiles{
		installScript: installScript,
		packageFile:   packageToInstall,
		cleanup: func() {
			_ = os.RemoveAll(extractDir)
		},
	}, nil
}

func extractPackage(extractDir string) (packageToInstall, installScript string, err error) {

	var (
		files []fs.DirEntry
	)

	files, err = moduleFiles.ReadDir(".")

	if err != nil {
		return "", "", fmt.Errorf("cannot read embedded files: %w", err)
	}

	packageFiles := make([]string, 0, 1)

	for _, f := range files {

		InstallLog.Printf("Extracting %s", f.Name())
		data, err := moduleFiles.ReadFile(f.Name())

		if err != nil {
			return "", "", fmt.Errorf("cannot read embedded file %s: %w", f.Name(), err)
		}

		dest, _ := filepath.Abs(filepath.Join(extractDir, f.Name()))

		if filepath.Ext(f.Name()) == ".nupkg" {
			packageFiles = append(packageFiles, dest)
		}

		if err := os.WriteFile(dest, data, 0600); err != nil {
			return "", "", fmt.Errorf("cannot write file %s: %w", dest, err)
		}
	}

	installScript, _ = filepath.Abs(filepath.Join(extractDir, "install-module.ps1"))

	if !fileExists(installScript) {
		return "", "", errors.New("missing script install-module.ps1 (packaging error)")
	}

	packageToInstall = getLatestPackage(packageFiles)
	if packageToInstall == "" {
		return "", "", errors.New("cannot find any PowerShell module to install (packaging error)")
	}

	return packageToInstall, installScript, nil
}

var semverRx = regexp.MustCompile(`((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?).nupkg$`)

func getLatestPackage(packages []string) string {

	switch len(packages) {
	case 0:
		return ""
	case 1:
		return packages[0]
	}

	// Extract versions to a new list
	versions := make([]*semver.Version, 0, len(packages))

	for _, p := range packages {
		if m := semverRx.FindStringSubmatch(p); len(m) > 0 {
			if v, err := semver.NewVersion(m[1]); err == nil {
				versions = append(versions, v)
			}
		}
	}

	var highestVersion string

	switch len(versions) {
	case 0:
		return ""
	case 1:
		highestVersion = versions[0].String()
	default:
		semver.Sort(versions)
		highestVersion = versions[len(versions)-1].String()
	}

	result := From(packages).
		Where(func(i any) bool {
			return strings.Contains(i.(string), highestVersion)
		}).First()

	if result != nil {
		return result.(string)
	}

	return ""
}

func runPowershell(scriptFile string, args ...string) error {

	psArgs := []string{
		"-NoProfile",
		"-NonInteractive",
		"-File",
		scriptFile,
	}

	if len(args) > 0 {
		psArgs = append(psArgs, args...)
	}

	//nolint: noctx // Command is long running - exit time is non-deterministic
	cmd := exec.Command("PowerShell.exe", psArgs...)

	// Stream to console and capture simultaneously
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func fileExists(path string) bool {

	if _, err := os.Stat(path); err != nil {
		return false
	}

	return true
}
