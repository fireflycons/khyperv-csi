function Dismount-Disk {
    <#
        .SYNOPSIS
            Dismounts a VHD from a Virtual Machine

        .DESCRIPTION
            Dismounts the VHD identified by DiskPath from the given Virtual Machine.
            If the disk doesn't exist, the function is a NOOP.

        .PARAMETER VMName
            Name of a virtual machine

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

    foreach ($controller in $controllers) {
        $drive = $controller.Drives | Where-Object { $_.Path -eq $DiskPath }

        if ($drive) {
            Remove-VMHardDiskDrive -VMName $vm.Name -ControllerType SCSI -ControllerNumber $drive.ControllerNumber -ControllerLocation $drive.ControllerLocation
            return
        }
    }
}

