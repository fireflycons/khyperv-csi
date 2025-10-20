function Get-DisksImpl{
    <#
        .SYNOPSIS
            List all defined PV volumes

        .DESCRIPTION
            Returns a JSON list of all volumes defined in the given PVStore directory

        .PARAMETER PVStore
            Directory where new PersistentVolume VHDs are stored

        .OUTPUTS
            [string] JSON list containing volume information
    #>

	param (
        [Parameter(Mandatory = $true)]
		[string]$PVStore,
        [int]$MaxEntries = 0,
        [string]$NextToken = ""
	)

    # Sometimes we get a windows short path here with a tilde (~) in the name.
    # Resolve it to the full path to avoid issues later on. 
    $PVStore = (Get-Item -Path $PVStore).FullName

    try {
    	$storeDisks = Get-ChildItem -Path $PVStore -Filter "*.vhd*" | Get-VHD
    }
    catch {
        throw "INVALID_ARGUMENT : " + $_.Exception.Message
    }

    $attachedDisks = Get-VM |
    ForEach-Object {
        Get-VMScsiController -VMName $_.Name |
        ForEach-Object {
            $_.Drives |
                Where-Object {
                    # Filter out all attached disks that that don't reside in the PVStore.
                    # These will be disks for VMs that are not managed by the CSI.
                    $PVStore -eq (Split-Path -Path $_.Path -Parent)
                }
        }
    }

    $unattachedDisks = if ($storeDisks -and -not $attachedDisks) {
        $storeDisks
    } elseif ($attachedDisks -and $storeDisks ) {
        Compare-Object -ReferenceObject $storeDisks -DifferenceObject $attachedDisks -Property Path -PassThru
    } else {
        @()
    }

    $allVolumes = Invoke-Command -ScriptBlock {

        $unattachedDisks |
        ForEach-Object {
            [PSCustomObject]@{
                DiskIdentifier = $_.DiskIdentifier
                Name = (Get-DiskName -Path $_.Path)
                Size = $_.Size
                Path = $_.Path
                Host = $null
            }
        }

        $attachedDisks |
        ForEach-Object {
            $attachedDisk = $_
            $vhd = Get-VHD -Path $attachedDisk.Path
            #$disk = $storeDisks | Where-Object { $_.Path -eq $attachedDisk.Path }
            $hostId = $attachedDisk.VMId
            [PSCustomObject]@{
                DiskIdentifier = $vhd.DiskIdentifier
                Name = (Get-DiskName -Path $_.Path)
                Size = $vhd.Size
                Path = $vhd.Path
                Host = $hostId
            }
        }
    }

    $allVolumesCount = ($allVolumes | Measure-Object).Count

    if ($allVolumesCount -eq 1) {
        $allVolumes = @(,$allVolumes)
    } elseif ($allVolumesCount -eq 0) {
        $allVolumes = @()
    }

    if ($MaxEntries -eq 0 -and $NextToken -eq "") {
        # Fast return
        return [PSCustomObject]@{
            VHDs = $allVolumes
            NextToken = ""
        }
    }

    $offset = 0

    if ($NextToken -match '^\d+$') {
        $offset = [int]$NextToken
    }

    $max = if ($MaxEntries -gt 0) {
        $MaxEntries
    } else {
        $allVolumesCount
    }

    $end = [Math]::Min($offset + $max, $allVolumesCount)

    if (-not $allVolumes) {
        return [PSCustomObject]@{
            VHDs = @();
            NextToken = ""
        }
    }

    $pagedVolumes = $allVolumes[$offset..($end-1)]
    $newNextToken = if ($end -lt $allVolumesCount) { $end.ToString() } else { "" }

    [PSCustomObject]@{
        VHDs = $pagedVolumes
        NextToken = $newNextToken
    }
}