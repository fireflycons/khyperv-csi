//go:build windows

package psmodule

import "embed"

//go:embed khyperv-csi.1.0.0.nupkg
//go:embed install-module.ps1
var moduleFiles embed.FS
