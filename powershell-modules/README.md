# powershell-modules

This directory contains a PowerShell module `khyperv-csi` which wraps Microsoft's `Hyper-V` module to perform the heavy lifting.

Given that this will be installed on the host that is running the Hyper-V service that is hosting your cluster, the `Hyper-V` module should be installed by default as part of the Hyper-V installation. You can assert its presence with the following command

```powershell
Get-Module -ListAvailable | Where-Object { $_.Name -eq 'Hyper-V' }
```

All the commands in the module that produce output produce that output as JSON such that it may be easily consumed by the Go wrapper.