function Get-Attachments {

    param (
        [Parameter(Mandatory = $true, ParameterSetName = 'ById')]
        [string]$VMId,

        [Parameter(Mandatory = $true, ParameterSetName = 'ByName')]
        [string]$VMName
    )

    try {
        if ($PSCmdlet.ParameterSetName -eq 'ByID') {
            $VMName = Get-VM -Id $VMId -ErrorAction Stop | Select-Object -ExpandProperty Name
        } else {
            $VMId = (Get-VM -Name $VMName -ErrorAction Stop).Id.Guid
        }
    } catch {
        throw "NOT_FOUND : " + $_.Exception.Message
    }

    $attachments = Get-VMScsiController -VMname $VMName |
        Select-Object -ExpandProperty Drives |
        ForEach-Object {
            $vhd = Get-VHD -Path $_.Path
            [PSCustomObject]@{
                DiskIdentifier = $vhd.DiskIdentifier
                Name = (Get-DiskName -Path $_.Path)
                Size = $vhd.Size
                Path = $_.Path
                Host = $VMId
            }
        }

    $result = switch ($attachments | Measure-Object | Select-Object -ExpandProperty Count) {
        0 {
            [PSCustomObject]@{
                VHDs = @()
                NextToken = ""
            }
        }

        1 {
            [PSCustomObject]@{
                VHDs = @($attachments)
                NextToken = ""
            }
        }

        Default {
            [PSCustomObject]@{
                VHDs = $attachments
                NextToken = ""
            }
        }
    }

    $result | ConvertTo-Json -Compress
}