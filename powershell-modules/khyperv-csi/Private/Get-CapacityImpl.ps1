function Get-CapacityImpl {
    <#
        .SYNOPSIS
            Gets available capacity for creating new VHDs

        .DESCRIPTION
            Get the available capacity for creating new VHDs.
            Capacity is worked out from free space on the drive then accounting
            for any space not yet claimed by dynamic disks.

        .PARAMETER PVStore
            Directory where new PersistentVolume VHDs are stored

        .OUTPUTS
            [int] Available free space
    #>
    param (
        [string]$PVStore
    )

    try {
        $PVStore = (Resolve-Path -Path $PVStore).Path
    }
    catch {
        throw "INVALID_ARGUMENT : " + $_.Exception.Message
    }

    $freeBytes = Get-Item $PVStore | Select-Object -ExpandProperty PSDrive | Select-Object -ExpandProperty Free

    Get-ChildItem -Path $PVStore -Filter "*.vhd*" |
    ForEach-Object {
        $disk = Get-VHD -Path $_.FullName

        if ($disk.VhdType -eq 'Dynamic') {
            $freeBytes -= ($disk.Size - $disk.FileSize)
        }
    }

    $freeBytes
}

