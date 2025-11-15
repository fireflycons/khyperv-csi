Write-Output ((
    Get-ChildItem -File -Path .\internal\logging\ -Recurse -Filter *.go |
        Where-Object {
            $_.Name -notlike 'linux*'
        } |
        ForEach-Object {
            (Resolve-Path -Relative $_.FullName).Replace("\", "/")
        }
) -join " ")