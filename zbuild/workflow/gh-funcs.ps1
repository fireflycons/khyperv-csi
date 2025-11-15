function Write-Log {
    param (
        [string]$Message
    )

    Write-Host ([Datetime]::Now.ToString("yyyy/MM/dd HH:mm:ss")) "-" $Message
}

function Get-Version {

    if ($env:GITHUB_REF_NAME -match 'v\d+\.\d+\.\d+') {
        $env:GITHUB_REF_NAME -replace '^[a-zA-Z]+', ''
    } else {
        Get-Content '.\_VERSION .txt' | Select-Object -First 1
    }
}