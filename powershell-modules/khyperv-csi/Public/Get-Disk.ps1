function Get-Disk {

    [CmdletBinding()]
    param (
        [Parameter(ParameterSetName = 'ByID')]
        [string]$Id,

        [Parameter(ParameterSetName = 'ByName')]
        [string]$Name,

        [Parameter(Mandatory, ParameterSetName = 'ByName')]
        [Parameter(Mandatory, ParameterSetName = 'ByID')]
        [string]$PVStore,

        [Parameter(ParameterSetName = 'ByPath')]
        [string]$Path
    )

    if ($PSCmdlet.ParameterSetName -ne "ByPath") {
        try {
            $PVStore = (Resolve-Path -Path $PVStore).Path
        }
        catch {
            throw "INVALID_ARGUMENT : " + $_.Exception.Message
        }
    }

    switch ($PSCmdlet.ParameterSetName) {
        'ByName' {
            try {
                $vhd = Get-ChildItem -Path $PVStore -Filter "${Name};*.vhd*" | Get-VHD
            }
            catch {
                throw "INVALID_ARGUMENT : " + $_.Exception.Message
            }

            if (($vhd | Measure-Object | Select-Object -ExpandProperty Count -First 1) -eq 0) {
                $vhd = $null
            } elseif (($vhd | Measure-Object | Select-Object -ExpandProperty Count -First 1) -gt 1) {
                throw "INTERNAL : Multiple volumes with name '$Name' found."
            }
            if (-not $vhd) {
                throw "NOT_FOUND : Volume with name '$Name' not found."
            }
        }
        'ById' {
            try {
                $vhd = Get-ChildItem -Path $PVStore -Filter "*;${Id}.vhd*" | Get-VHD | Where-Object { $_.DiskIdentifier -eq $Id }
            }
            catch {
                throw "INVALID_ARGUMENT : " + $_.Exception.Message
            }

            if (-not $vhd) {
                throw "NOT_FOUND : Volume with id '$Id' not found."
            }
        }
        'ByPath' {
            if (-not (Test-Path -Path $Path)) {
                throw "NOT_FOUND : Volume with path '$Path' not found."
            }
            try {
                $vhd = Get-VHD -Path $Path
            } catch {
                throw "NOT_FOUND : " + $_.Exception.Message
            }
        }
        Default {
            throw "INTERNAL : Invalid parameter set."
        }
    }

    $vhd | ConvertTo-Json -Compress
}