//go:build windows

package psmodule

import (
	"fmt"
)

func RemoveModule() error {

	extracted, err := extractEmbeddedFiles()

	if err != nil {
		return fmt.Errorf("remove-module: %w", err)
	}

	defer extracted.cleanup()

	InstallLog.Println("Removing PowerShell module khyperv-csi...")

	if err := runPowershell(extracted.installScript, "-Remove"); err != nil {
		return fmt.Errorf("remove-module: Error removing PowerShell module: %v", err.Error())
	}

	return nil
}
