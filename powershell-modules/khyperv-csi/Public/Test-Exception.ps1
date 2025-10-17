function Test-Exception {
    <#
    .SYNOPSIS
        Throws a canonical CSI error with a custom message.

    .DESCRIPTION
        This function throws a canonical CSI error with a custom message.
        It is useful for testing error handling in CSI clients.

    .PARAMETER CanonicalCsiError
        The canonical CSI error to throw. This should be one of the standard CSI error codes.
    #>
    param (
        [string]$CanonicalCsiError
    )

    throw "$CanonicalCsiError : An error occurred while processing your request."
}