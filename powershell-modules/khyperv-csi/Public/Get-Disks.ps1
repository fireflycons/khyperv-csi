function Get-Disks{
    <#
        .SYNOPSIS
            List all defined PV volumes

        .DESCRIPTION
            Returns a JSON list of all volumes defined in the given PVStrore directory

        .PARAMETER PVStore
            Directory where new PersistentVolume VHDs are stored

        .OUTPUTS
            [string] JSON object containing volume information
    #>

	param (
        [Parameter(Mandatory = $true)]
		[string]$PVStore,
        [int]$MaxEntries,
        [string]$NextToken
	)

    $result = (Get-DisksImpl -PVStore $PVStore -MaxEntries $MaxEntries -NextToken $NextToken | ConvertTo-Json -Compress).Trim()

    switch ($result) {
        {$_ -eq ""}             { "{}" }
        Default                 { $result }
    }
}

