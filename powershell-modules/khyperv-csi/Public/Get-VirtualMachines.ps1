function Get-VirtualMachines {

    <#
        .SYNOPSIS
        Get all VMs
    #>

    try {
        $vms = Get-VM | Select-Object Name, ID, Path, Generation

        switch ($vms | Measure-Object | Select-Object -ExpandProperty Count) {
            0 { $vms = @() }
            1 { $vms = @(,$vms) }
        }
        [PSCustomObject]@{
            VMs = $vms
        } | ConvertTo-Json -Compress

    } catch {

        throw "INTERNAL : cannot list VMs : " +  + $_.Exception.Message
    }
}