function Mount-Disk {
    <#
        .SYNOPSIS
            Mounts a VHD to a Virtual Machine

        .DESCRIPTION
            Mount a VHD to the first available slot on the given VM's SCSI interface

        .PARAMETER VMName

        .PARAMETER VMId
            ID of a virtual machine

        .PARAMETER DiskPath
            Path to a VHD disk file
    #>
    param (
        [Parameter(Mandatory, ParameterSetName = "ByName")]
        [string]$VMName,

        [Parameter(Mandatory, ParameterSetName = "ById")]
        [string]$VMId,

        [Parameter(Mandatory = $true)]
        [string]$DiskPath
    )

    $vm = switch ($PSCmdlet.ParameterSetName) {
        'ByID' {
            Get-VM -Id $VMId -ErrorAction SilentlyContinue
         }
        'ByName' {
            Get-VM -Name $VMName -ErrorAction SilentlyContinue
        }
    }

    if (-not $vm) {
        throw "NOT_FOUND : VM does not exist"
    }

    $controllers = $vm | Get-VMScsiController

    if (-not $controllers) {
        # Max supported volumes is zero in this case.
        throw "RESOURCE_EXHAUSTED : No SCSI interfaces"
    }

    $allDrives = $controllers | Select-Object -ExpandProperty Drives

    if (-not $allDrives) {

        # No disks added yet
        $controller = $controllers | Select-Object -First 1
        $controllerLocation = 0

    } else {

        # Check whether this disk is already mounted. If so, just return the disk
        $mounted = $allDrives |
            Where-Object { $_.Path -eq $DiskPath}

        if ($mounted) {
            $mounted | ConvertTo-Json -Compress
            return
        }

        # First controller with a free location
        $controller = $controllers |
            Where-Object { ($_.Drives | Measure-Object | Select-Object -ExpandProperty Count) -lt $script:MaxVolumesPerController } |
            Select-Object -First 1

        if (-not $controller) {
             throw "RESOURCE_EXHAUSTED : No free slots"
        }

        $controllerLocation = if (-not $controller.Drives) {
            0
        } else {
            $locs = $controller.Drives.ControllerLocation | Sort-Object
            (0..63) | Where-Object { $_ -notin $locs } | Select-Object -First 1
        }
    }

    # If we get here, then we have a controller and location
    try {
        $controller | Add-VMHardDiskDrive -Passthru -Path $DiskPath -ControllerLocation $controllerLocation | ConvertTo-Json -Compress
    } catch {
        if ($_.Exception.Message.Contains("The disk is already connected")) {
            throw "FAILED_PRECONDITION : " + $_.Exception.Message
        }
        throw "INTERNAL : " + $_.Exception.Message
    }
}