param(
    [switch]$DryRun,
    [switch]$SkipPlaywright,
    [switch]$SkipVite
)

$ErrorActionPreference = "Stop"
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)

function Get-CommandPreview {
    param([string]$CommandLine)

    if ([string]::IsNullOrWhiteSpace($CommandLine)) {
        return ""
    }

    $compact = $CommandLine -replace "\s+", " "
    if ($compact.Length -le 150) {
        return $compact
    }

    return $compact.Substring(0, 147) + "..."
}

function Get-ProcessSnapshot {
    $snapshot = @{}
    Get-Process -ErrorAction SilentlyContinue | ForEach-Object {
        $snapshot[[int]$_.Id] = $_
    }
    return $snapshot
}

function Show-Targets {
    param(
        [string]$Title,
        [object[]]$Targets
    )

    Write-Host $Title
    if (-not $Targets -or $Targets.Count -eq 0) {
        Write-Host "  none"
        return
    }

    $processSnapshot = Get-ProcessSnapshot
    $Targets | Sort-Object ProcessId -Unique | ForEach-Object {
        $processInfo = $processSnapshot[[int]$_.ProcessId]
        [pscustomobject]@{
            PID   = [int]$_.ProcessId
            PPID  = [int]$_.ParentProcessId
            Name  = $_.Name
            WS_MB = if ($processInfo) { [math]::Round($processInfo.WorkingSet64 / 1MB, 1) } else { $null }
            Cmd   = Get-CommandPreview $_.CommandLine
        }
    } | Format-Table -Wrap -AutoSize
}

function Stop-TargetProcesses {
    param([object[]]$Targets)

    if (-not $Targets -or $Targets.Count -eq 0) {
        return
    }

    foreach ($target in ($Targets | Sort-Object ProcessId -Descending -Unique)) {
        if ($DryRun) {
            continue
        }

        Stop-Process -Id ([int]$target.ProcessId) -Force -ErrorAction SilentlyContinue
    }
}

function Get-PlaywrightTargets {
    $targets = @()
    $targets += Get-CimInstance Win32_Process -Filter "name = 'node.exe'" |
        Where-Object { $_.CommandLine -like "*playwright-core*lib*entry*cliDaemon.js*" }
    $targets += Get-CimInstance Win32_Process -Filter "name = 'chrome.exe'" |
        Where-Object { $_.CommandLine -like "*playwright_chromiumdev_profile*" }

    return @($targets | Where-Object { $_ -and $_.ProcessId } | Sort-Object ProcessId -Unique)
}

function Get-DuplicateViteTargets {
    $listeners = @{}
    Get-NetTCPConnection -State Listen -ErrorAction SilentlyContinue | ForEach-Object {
        $listeners[[int]$_.LocalPort] = [int]$_.OwningProcess
    }

    $targets = @()
    $viteNodes = Get-CimInstance Win32_Process -Filter "name = 'node.exe'" |
        Where-Object { $_.CommandLine -like "*vite*bin*vite.js*" }

    foreach ($process in $viteNodes) {
        $port = $null
        if ($process.CommandLine -match "--port\s+(\d+)") {
            $port = [int]$Matches[1]
        }

        if ($null -eq $port) {
            continue
        }

        if ($listeners.ContainsKey($port) -and $listeners[$port] -ne [int]$process.ProcessId) {
            $targets += $process

            $parent = Get-CimInstance Win32_Process -Filter "ProcessId = $($process.ParentProcessId)" -ErrorAction SilentlyContinue
            if ($parent -and $parent.Name -eq "cmd.exe" -and $parent.CommandLine -like "*vite*") {
                $targets += $parent
            }
        }
    }

    return @($targets | Where-Object { $_ -and $_.ProcessId } | Sort-Object ProcessId -Unique)
}

$allTargets = @()

if (-not $SkipPlaywright) {
    $playwrightTargets = Get-PlaywrightTargets
    Show-Targets "Playwright cleanup targets:" $playwrightTargets
    $allTargets += $playwrightTargets
}

if (-not $SkipVite) {
    $viteTargets = Get-DuplicateViteTargets
    Show-Targets "Duplicate Vite cleanup targets:" $viteTargets
    $allTargets += $viteTargets
}

$allTargets = @($allTargets | Where-Object { $_ -and $_.ProcessId } | Sort-Object ProcessId -Unique)

if ($DryRun) {
    Write-Host "Dry run only. No process was stopped."
} else {
    Stop-TargetProcesses $allTargets
    Start-Sleep -Seconds 1
}

$remainingPlaywright = @()
$remainingVite = @()

if (-not $SkipPlaywright) {
    $remainingPlaywright = Get-PlaywrightTargets
}
if (-not $SkipVite) {
    $remainingVite = Get-DuplicateViteTargets
}

$os = Get-CimInstance Win32_OperatingSystem
$cpu = Get-CimInstance Win32_Processor | Select-Object -First 1

Write-Host ""
Write-Host "Summary:"
Write-Host ("  stopped candidates: {0}" -f $allTargets.Count)
Write-Host ("  remaining Playwright targets: {0}" -f $remainingPlaywright.Count)
Write-Host ("  remaining duplicate Vite targets: {0}" -f $remainingVite.Count)
Write-Host ("  free RAM: {0} GB" -f ([math]::Round($os.FreePhysicalMemory / 1MB, 1)))
Write-Host ("  CPU load: {0}%" -f $cpu.LoadPercentage)
