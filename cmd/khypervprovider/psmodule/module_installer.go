//go:build windows

package psmodule

import (
	"fmt"
	"log"
	"path/filepath"
)

func InstallModule() error {

	extracted, err := extractEmbeddedFiles()

	if err != nil {
		return fmt.Errorf("install-module: %w", err)
	}

	defer extracted.cleanup()

	log.Printf("Installing PowerShell module %s...", filepath.Base(extracted.packageFile))

	if err := runPowershell(extracted.installScript, "-Package", extracted.packageFile); err != nil {
		return fmt.Errorf("install-module: Error installing PowerShell module: %v", err.Error())
	}

	return nil
}
