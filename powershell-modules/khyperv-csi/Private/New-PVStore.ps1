function New-PVStore {

    # select volume with most space
    $volume = Get-Volume |
        Where-Object { $_.DriveLetter  -and $_.DriveType -eq 'Fixed'} |
        Sort-Object -Descending SizeRemaining |
        Select-Object -First 1

    $pvstore = $volume.DriveLetter + ":\" + $script:PVStoreName

    if (-not (Test-Path -PathType Container -Path $pvstore)) {
        try {
            New-Item -ItemType Directory -Path $pvstore
        } catch {
            throw "INTERNAL : " + $_.Exception.Message
        }
    }

    $pvstore
}