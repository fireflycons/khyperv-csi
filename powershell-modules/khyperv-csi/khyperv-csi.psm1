$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# Get public and private function definition files.
$Public = @( Get-ChildItem -Path ([IO.Path]::Combine($PSScriptRoot, "Public", "*.ps1")) -Recurse -ErrorAction SilentlyContinue )
$Private = @( Get-ChildItem -Path ([IO.Path]::Combine($PSScriptRoot, "Private", "*.ps1")) -Recurse -ErrorAction SilentlyContinue )

# Dot source the files
foreach ($import in @($Public + $Private))
{
    try
    {
        . $import.FullName
    }
    catch
    {
        Write-Error -Message "Failed to import function $($import.FullName): $_"
    }
}

$script:PVStoreName = "Kubernetes Persistent Volumes"
$script:MinVolumeSize = 5MB
$script:MaxVolumesPerController = 64