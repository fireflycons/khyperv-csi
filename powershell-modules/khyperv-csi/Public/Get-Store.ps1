function Get-Store {

    try {
        $storeVol = Get-Volume |
            Where-Object {
                $_.DriveLetter  -and $_.DriveType -eq 'Fixed' -and (Test-Path -PathType Container -Path ($_.DriveLetter + ":\" + $script:PVStoreName))
            }

        $store = Invoke-Command -ScriptBlock {
            if ($storeVol) {
                $storeVol.DriveLetter + ":\" + $script:PVStoreName
            } else {
                New-PVStore
            }
        }
        @{PVStore = $store} | ConvertTo-Json -Compress
    }
    catch {
        throw "INTERNAL : " + $_.Exception.Message
    }
}