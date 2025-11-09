function Resize-Disk {

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
        [string]$Path,

        [Parameter(Mandatory=$true)]
        [System.Int64]$Size
    )

    $getDiskParams = @{}
    foreach ($key in $PSBoundParameters.Keys) {
        if ($key -notin @('Size')) {
            $getDiskParams[$key] = $PSBoundParameters[$key]
        }
    }

    # Will throw if the disk can't be retrieved
    $vhd = Get-Disk @getDiskParams

    # Idempotency
    if ($vhd.Size -ge $Size) {
        $vhd | ConvertTo-Json -Compress
        return
    }

    # Initial capacity check
    $free = Get-CapacityImpl -PVStore $PVStore

    if ($Size - $vhd.Size -gt $free - $script:MinFreeSpace) {
        throw "OUT_OF_RANGE : New size exceeds minimum free space limit in volume store"
    }

    try {
        Resize-VHD -Path $vhd.Path -SizeBytes $Size -ErrorAction Stop | Out-Null
        $resizedVhd = Get-VHD -Path $vhd.Path
    }
    catch [Microsoft.HyperV.PowerShell.VirtualizationException] {
        $err = $_.Exception
        $msg = $err.Message
        $hresult = ('0x{0:X8}' -f ($err.HResult -band 0xFFFFFFFF))

        switch ($hresult) {
            '0x80070070' {  # ERROR_DISK_FULL
                throw "OUT_OF_RANGE : New size exceeds minimum free space limit in volume store ($hresult)"
            }
            '0x80070057' {  # E_INVALIDARG
                throw "INVALID_ARGUMENT : invalid size or argument. ($hresult)"
            }
            '0x800700DF' {  # ERROR_FILE_TOO_LARGE
                throw "OUT_OF_RANGE : file system limitation (file too large). ($hresult)"
            }
            default {
                throw "INVALID_ARGUMENT : Resize failed with virtualization exception: $msg ($hresult)"
            }
        }
    }
    catch {
        throw "INTERNAL : Unexpected error: $($_.Exception.Message)"
    }

    $resizedVhd | ConvertTo-Json -Compress
}

