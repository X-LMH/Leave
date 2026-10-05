param(
    [string]$DeviceSerial = '',
    [string]$AdbPath = ''
)

$ErrorActionPreference = 'Stop'

try {
    if ($AdbPath -eq '') {
        $adbCommand = Get-Command adb -CommandType Application -ErrorAction SilentlyContinue
        if ($null -ne $adbCommand) {
            $AdbPath = $adbCommand.Source
        } else {
            $AdbPath = 'D:\Code\IDE\HBuilderX\plugins\launcher-tools\tools\adbs\adb.exe'
        }
    }
    if (-not (Test-Path -LiteralPath $AdbPath -PathType Leaf)) {
        throw 'ADB was not found. Pass -AdbPath with the path to adb.exe.'
    }

    $deviceLines = @(& $AdbPath devices)
    if ($LASTEXITCODE -ne 0) { throw 'Could not list ADB devices.' }
    $devices = @(
        foreach ($line in $deviceLines) {
            if ($line -match '^(\S+)\s+device$') { $Matches[1] }
        }
    )
    if ($DeviceSerial -eq '') {
        # Never silently pick the first device when several are connected.
        if ($devices.Count -ne 1) {
            $deviceLines | ForEach-Object { Write-Host $_ }
            throw 'Connect and authorize exactly one device, or pass -DeviceSerial.'
        }
        $DeviceSerial = $devices[0]
    } elseif ($devices -notcontains $DeviceSerial) {
        throw "Device $DeviceSerial is not connected and authorized."
    }

    $mappingLines = @(& $AdbPath -s $DeviceSerial reverse --list)
    if ($LASTEXITCODE -ne 0) { throw 'Could not read existing port mappings.' }
    $mappings = @{}
    foreach ($line in $mappingLines) {
        if ($line -match '\s(tcp:\d+)\s+(\S+)\s*$') {
            $mappings[$Matches[1]] = $Matches[2]
        }
    }

    # Check both ports before changing anything; preserve unrelated mappings.
    foreach ($port in @(8080, 8081)) {
        $endpoint = "tcp:$port"
        if ($mappings.ContainsKey($endpoint) -and $mappings[$endpoint] -ne $endpoint) {
            throw "$endpoint already maps to $($mappings[$endpoint]); no mappings were changed."
        }
        $client = New-Object System.Net.Sockets.TcpClient
        try {
            $connection = $client.ConnectAsync('127.0.0.1', $port)
            if (-not $connection.Wait(2000)) { throw "Port $port timed out." }
        } finally {
            $client.Dispose()
        }
    }

    foreach ($port in @(8080, 8081)) {
        $endpoint = "tcp:$port"
        if (-not $mappings.ContainsKey($endpoint)) {
            & $AdbPath -s $DeviceSerial reverse $endpoint $endpoint
            if ($LASTEXITCODE -ne 0) { throw "Could not forward $endpoint. Run this script again after reconnecting." }
        }
    }

    $verifiedLines = @(& $AdbPath -s $DeviceSerial reverse --list)
    if ($LASTEXITCODE -ne 0) { throw 'Could not verify port mappings.' }
    foreach ($port in @(8080, 8081)) {
        if (-not ($verifiedLines -match "\stcp:$port\s+tcp:$port\s*$")) {
            throw "Port $port mapping is missing. Run this script again after reconnecting."
        }
    }
    Write-Host "USB forwarding is ready for device $DeviceSerial." -ForegroundColor Green
    Write-Host '8080: Go backend; 8081: avatar server. Reopen or retry the App.'
    Write-Host 'Run this script again after reconnecting USB or restarting ADB.'
    exit 0
} catch {
    Write-Host "Failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host 'Keep the Go backend and Nginx avatar server running; allow USB debugging on the phone.'
    exit 1
}
