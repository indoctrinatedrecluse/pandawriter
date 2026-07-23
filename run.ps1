[CmdletBinding()]
param(
    [switch]$NoLaunch
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSCommandPath
$buildDirectory = Join-Path $projectRoot 'build\bin'
$targetBinary = Join-Path $buildDirectory 'pandawriter.exe'

Set-Location -LiteralPath $projectRoot

if (Test-Path -LiteralPath $buildDirectory) {
    $resolvedBuildDirectory = (Resolve-Path -LiteralPath $buildDirectory).Path
    $expectedBuildDirectory = [System.IO.Path]::GetFullPath($buildDirectory)
    if ($resolvedBuildDirectory -ne $expectedBuildDirectory) {
        throw "Refusing to clear an unexpected build directory: $resolvedBuildDirectory"
    }

    Get-ChildItem -LiteralPath $resolvedBuildDirectory -Force | Remove-Item -Recurse -Force
}

Write-Host 'Building PandaWriter...'
$wails = Get-Command wails -ErrorAction SilentlyContinue
if ($wails) {
    & $wails.Source build
} else {
    Write-Host 'Wails CLI was not found on PATH; using the project-pinned Wails CLI.'
    go run github.com/wailsapp/wails/v2/cmd/wails@v2.13.0 build
}

if ($LASTEXITCODE -ne 0) {
    throw "Wails build failed with exit code $LASTEXITCODE."
}
if (-not (Test-Path -LiteralPath $targetBinary -PathType Leaf)) {
    throw "Expected build output was not created: $targetBinary"
}

if (-not $NoLaunch) {
    Write-Host 'Launching PandaWriter...'
    Start-Process -FilePath $targetBinary -WorkingDirectory $buildDirectory
}
