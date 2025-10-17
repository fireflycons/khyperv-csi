function Remove-Disk {

    [CmdletBinding(DefaultParameterSetName = 'ByID')]
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

        if (-not (Test-Path -Path $PVStore -PathType Container)) {
            # Same as disk not found, due to idempotency
            return
        }
    }

    $fullPath = switch ($PSCmdlet.ParameterSetName) {

        'ByPath' { $Path }

        'ByID' {
            Get-ChildItem -Path $PVStore -Filter "*.vhd*" |
                Where-Object {
                    $_.Name -ilike "*${Id}.vhd*"
                } |
                Select-Object -ExpandProperty FullName
        }

        'ByName' {
            Get-ChildItem -Path $PVStore -Filter "*.vhd*" |
                Where-Object {
                    $_.Name.StartsWith($Name+";")
                } |
                Select-Object -ExpandProperty FullName
        }
    }

    if ($fullPath -is [array]) {
        throw "INTERNAL : Duplicate disks found"
    }

    # If the file doesn't exist, it's not an error due to idempotency
    if (-not $fullPath -or -not (Test-Path -PathType Leaf -Path $fullPath)) {
        return
    }

    $vhd = Get-VHD -Path $fullPath

    # Get-VHD only sets Attached=True if the VM that has the attachment is running.
    if ($vhd | Select-Object -ExpandProperty Attached) {
        throw "FAILED_PRECONDITION : Disk is attached"
    }

    # Long route
    $allDisks = Get-DisksImpl -PVStore $PVStore

    if ($allDisks.VHDs | Where-Object { $Id -eq $_.DiskIdentifier -and  $null -ne $_.Host}) {
        throw "FAILED_PRECONDITION : Disk is attached"
    }

    Remove-Item -Path $fullPath | Out-Null
}