$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$settings = Get-Content (Join-Path $projectRoot '.private/dev-ssh.json') -Raw | ConvertFrom-Json
$key = Join-Path $projectRoot $settings.identityFile
# Make may launch 32-bit PowerShell; Sysnative reaches the 64-bit OpenSSH client.
$systemDirectory = if ([Environment]::Is64BitProcess) { 'System32' } else { 'Sysnative' }
$ssh = Join-Path $env:WINDIR "$systemDirectory/OpenSSH/ssh.exe"
$air = Get-Command air -ErrorAction SilentlyContinue
$airPath = if ($air) { $air.Source } else { Join-Path $env:USERPROFILE 'go/bin/air.exe' }
$sshProcess = $null
$exitCode = 0
Push-Location (Join-Path $projectRoot 'backend')
try {
    if (-not (Test-Path -LiteralPath $airPath)) { throw 'Air was not found.' }
    $ports = [Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners().Port
    if ($ports -contains 16379 -or $ports -contains 13306) {
        throw 'Port 16379 or 13306 is already in use.'
    }
    $sshArguments = @(
        '-N', '-T', '-i', ('"{0}"' -f $key),
        '-L', '127.0.0.1:16379:127.0.0.1:6379',
        '-L', '127.0.0.1:13306:127.0.0.1:3306',
        '-o', 'BatchMode=yes', '-o', 'StrictHostKeyChecking=yes',
        '-o', 'ExitOnForwardFailure=yes', '-o', 'ConnectTimeout=10',
        '-o', 'ServerAliveInterval=15', '-o', 'ServerAliveCountMax=3',
        $settings.host
    )
    $sshProcess = Start-Process $ssh -ArgumentList $sshArguments -PassThru -WindowStyle Hidden
    $deadline = [DateTime]::UtcNow.AddSeconds(12)
    do {
        if ($sshProcess.HasExited) { throw 'SSH tunnel failed. Check SSH connectivity and the key.' }
        if ([DateTime]::UtcNow -gt $deadline) { throw 'SSH tunnel startup timed out.' }
        Start-Sleep -Milliseconds 100
        $ports = [Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners().Port
        $ready = $ports -contains 16379 -and $ports -contains 13306
    } until ($ready)
    Write-Host 'Redis :16379 and MySQL :13306 tunnels ready. Starting Air...'
    & $airPath -c .air.toml
    $exitCode = $LASTEXITCODE
} catch {
    Write-Host $_.Exception.Message -ForegroundColor Red
    $exitCode = 1
} finally {
    if ($null -ne $sshProcess) {
        if (-not $sshProcess.HasExited) {
            $sshProcess.Kill()
            $sshProcess.WaitForExit()
        }
        $sshProcess.Dispose()
        Write-Host 'Redis and MySQL tunnels closed.'
    }
    Pop-Location
}
exit $exitCode
