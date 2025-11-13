function New-Disk {

    <#
        .SYNOPSIS
            Creates a new VHD

        .DESCRIPTION
            Creates a VHD of the given type.
            Disks are created as dynmamic so that the function returns quickly

        .PARAMETER Name
            Name of the disk, e.g. "PV1"

        .PARAMETER PVStore
            Directory to create disk in

        .PARAMETER Size
            Requiested size, in bytes

        .PARAMETER VHDType
            Type of VHD
    #>
    param (
        [Parameter(Mandatory = $true)]
        [string]$Name,

        [Parameter(Mandatory = $true)]
        [string]$PVStore,

        [Parameter(Mandatory = $true)]
        [System.Int64]$Size,

        [Parameter(Mandatory = $true)]
        [ValidateSet('.vhdx', '.vhd')]
        [string]$VHDType
    )

    try {
        $PVStore = (Resolve-Path -Path $PVStore).Path
    }
    catch {
        throw "INVALID_ARGUMENT : " + $_.Exception.Message
    }

    # Check for duplicate name
    $existingDisk = Get-ChildItem -Path $PVStore -Filter "*.vhd*" |
        Where-Object { $_.Name.StartsWith($Name+";")}

    if ($existingDisk) {
        $vhd = Get-VHD -Path $existingDisk.FullName

        if ($vhd.Size -eq $Size) {
            # Idempotency
            $vhd | ConvertTo-Json -Compress
            return
        }

        throw "ALREADY_EXISTS : Disk with name ${Name} alreay exists with different properties"
    }

    if ($Size -lt $script:MinVolumeSize) {
        $Size = $script:MinVolumeSize
    }

    $freeBytes = Get-CapacityImpl -PVStore $PVStore

    $consumedSize = Invoke-Command -ScriptBlock {
        if ((Get-Item -Path $PVStore).PSDrive.Name -eq 'C') {
            # If C drive, leave 5GB free
            $Size + $script:MinFreeSpace
        } else {
            # Add a meg to allow for bock sizing etc
            $Size + 1MB
        }
    }

    if ($freeBytes -lt $consumedSize) {
        throw "RESOURCE_EXHAUSTED : Insufficient storage"
    }

    $tempName = [Guid]::NewGuid().Guid
    $tempPath = Join-Path -Path $PVStore -ChildPath ($tempName + $VHDType)

    try {
        $disk = New-VHD -Path $tempPath -Dynamic -SizeBytes $Size

        # Rename the file to have the requested name and disk identifier as the filename
        # for faster lookup
        $newName = Join-Path -Path $PVStore -ChildPath ($name + ";" + $disk.DiskIdentifier + $VHDType)
        Move-Item -Path $tempPath -Destination $newName
        Get-VHD -Path $newName | ConvertTo-Json -Compress
    }
    catch {
        throw "INTERNAL : " + $_.Exception.Message
    }
}