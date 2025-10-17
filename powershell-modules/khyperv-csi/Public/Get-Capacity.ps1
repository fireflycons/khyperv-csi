function Get-Capacity {
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
            [string] JSON object containing single integer field "Free"
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$PVStore
    )


    @{"Free" = (Get-CapacityImpl -PVStore $PVStore)} | ConvertTo-Json
}

