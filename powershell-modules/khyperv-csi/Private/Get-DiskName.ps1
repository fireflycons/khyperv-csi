function Get-DiskName {
    param (
        [string]$Path
    )

    [IO.Path]::GetFileNameWithoutExtension($Path) -split ';' | Select-Object -First 1
}

