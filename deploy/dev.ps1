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
$sshErrorLog = Join-Path ([IO.Path]::GetTempPath()) ('leave-dev-ssh-{0}.log' -f [guid]::NewGuid())
$exitCode = 0

function Test-TunnelPort([int]$Port) {
    $client = New-Object Net.Sockets.TcpClient
    try {
        $connection = $client.ConnectAsync('127.0.0.1', $Port)
        if (-not $connection.Wait(200)) { return $false }
        return $client.Connected
    } catch {
        return $false
    } finally {
        $client.Dispose()
    }
}

Push-Location (Join-Path $projectRoot 'backend')
try {
    if (-not (Test-Path -LiteralPath $airPath)) { throw 'Air was not found.' }
    if ((Test-TunnelPort 16379) -or (Test-TunnelPort 13306)) {
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
    $sshProcess = Start-Process $ssh -ArgumentList $sshArguments -PassThru -WindowStyle Hidden -RedirectStandardError $sshErrorLog
    $deadline = [DateTime]::UtcNow.AddSeconds(12)
    do {
        $sshProcess.Refresh()
        if ($sshProcess.HasExited) { throw 'SSH tunnel failed. Check SSH connectivity and the key.' }
        $ready = (Test-TunnelPort 16379) -and (Test-TunnelPort 13306)
        if ($ready) { break }
        if ([DateTime]::UtcNow -gt $deadline) { throw 'SSH tunnel startup timed out.' }
        Start-Sleep -Milliseconds 100
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
        if ($exitCode -ne 0 -and (Test-Path -LiteralPath $sshErrorLog)) {
            Get-Content -LiteralPath $sshErrorLog | ForEach-Object { Write-Host $_ -ForegroundColor Red }
        }
        Write-Host 'Redis and MySQL tunnels closed.'
    }
    Remove-Item -LiteralPath $sshErrorLog -ErrorAction SilentlyContinue
    Pop-Location
}
exit $exitCode
