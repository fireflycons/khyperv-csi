function Get-VMId {

    param (
        [Parameter(Mandatory=$true)]
        [string]$VMName
    )

    $vm = Get-VM -Name $VMName -ErrorAction Stop

    @{ VMId = $vm.Id.Guid.ToString() } | ConvertTo-Json -Compress
}