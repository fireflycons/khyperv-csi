function New-TestVM {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Name,

        [Parameter(Mandatory=$true)]
        [string]$Path,

        [Parameter(Mandatory=$true)]
        [ValidateSet(1,2)]
        [int]$Generation
    )

    # Settings
    $memory = 512MB
    $processorCount = 1

    try {
        # Abort if VM already exists
        if (Get-VM -Name $Name -ErrorAction SilentlyContinue) {
            throw "A VM named '$Name' already exists. Choose a different name or remove the existing VM."
        }

        # Create the VM with no VHD and specified generation & memory
        New-VM -Name $Name -Path $Path -Generation $Generation -MemoryStartupBytes $memory -NoVHD | Out-Null

        # Ensure processor count
        Set-VMProcessor -VMName $Name -Count $processorCount

        # Remove any DVD drives (CD-ROM) if created by default
        Get-VMDvdDrive -VMName $Name -ErrorAction SilentlyContinue | ForEach-Object {
            Remove-VMDvdDrive -VMName $Name -ControllerNumber $_.ControllerNumber -ControllerLocation $_.ControllerLocation -Confirm:$false
        }

        # Ensure there's at least one SCSI controller
        $scsiControllers = Get-VMScsiController -VMName $Name -ErrorAction SilentlyContinue
        if (-not $scsiControllers) {
            Add-VMScsiController -VMName $Name | Out-Null
        }

        # Networking: attach to "Default Switch" if available, otherwise first available switch
        $switch = Get-VMSwitch -Name "Default Switch" -ErrorAction SilentlyContinue
        if (-not $switch) {
            $switch = Get-VMSwitch | Select-Object -First 1
        }

        if ($switch) {
            Add-VMNetworkAdapter -VMName $Name -SwitchName $switch.Name -Name "Network Adapter" | Out-Null
        }

        Get-VM -Name $Name | Select-Object Name, ID, Path, Generation | ConvertTo-Json -Compress
    }
    catch {
        throw "INTERNAL : ${_.Exception.Message} Failed to create VM '$Name': " + $_.Exception.Message
    }
}