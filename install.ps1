<#
.SYNOPSIS
Installs the Spring Boot CLI (springcli) on Windows.

.DESCRIPTION
This script downloads the latest release of springcli for Windows from GitHub,
places it in the user's home directory (~\.springcli\bin), and adds that
directory to the user's PATH environment variable if it's not already there.

.EXAMPLE
.\install.ps1
#>

$ErrorActionPreference = "Stop"

$Repo = "Dineshs737/springboot-cli"
$BinaryName = "springcli-windows-amd64.exe"
$InstallDir = Join-Path $HOME ".springcli\bin"
$ExecutablePath = Join-Path $InstallDir "springcli.exe"

Write-Host "Installing SpringCLI for Windows..." -ForegroundColor Cyan

# Ensure the installation directory exists
if (-not (Test-Path -Path $InstallDir)) {
    Write-Host "Creating installation directory at $InstallDir..."
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

# Fetch the latest release version from GitHub API
Write-Host "Fetching latest version from GitHub..."
try {
    $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    $LatestVersion = $ReleaseInfo.tag_name
}
catch {
    Write-Warning "Failed to fetch the latest release version. Falling back to specific tag or rate limit hit."
    $LatestVersion = "v1.0.0"
}

Write-Host "Downloading $BinaryName version $LatestVersion..."

# Construct the download URL
$DownloadUrl = "https://github.com/$Repo/releases/download/$LatestVersion/$BinaryName"

# Download the executable
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExecutablePath -UseBasicParsing
    Write-Host "Download complete." -ForegroundColor Green
}
catch {
    Write-Warning "Failed to download the binary from $DownloadUrl"
    
    # Fallback for local development/testing where releases aren't published
    $LocalBinary = Join-Path $PSScriptRoot "springcli.exe"
    if (Test-Path $LocalBinary) {
        Write-Host "Found local springcli.exe. Copying it to installation directory..." -ForegroundColor Yellow
        Copy-Item -Path $LocalBinary -Destination $ExecutablePath -Force
        Write-Host "Local copy complete." -ForegroundColor Green
    } else {
        Write-Error "Could not download the binary and local $LocalBinary was not found."
        exit 1
    }
}

# Check and update the PATH environment variable
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$NeedsPathUpdate = $true

if ($UserPath) {
    $PathEntries = $UserPath -split ";"
    foreach ($Entry in $PathEntries) {
        if ($Entry.TrimEnd('\') -eq $InstallDir.TrimEnd('\')) {
            $NeedsPathUpdate = $false
            break
        }
    }
}

if ($NeedsPathUpdate) {
    Write-Host "Adding $InstallDir to your PATH..."
    if ($UserPath) {
        $NewPath = $UserPath + ";" + $InstallDir
    } else {
        $NewPath = $InstallDir
    }
    
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Host "Successfully added to PATH!" -ForegroundColor Green
    Write-Host "Please restart your PowerShell terminal for the changes to take effect." -ForegroundColor Yellow
} else {
    Write-Host "The directory $InstallDir is already in your PATH."
}

Write-Host ""
Write-Host "SpringCLI installed successfully!" -ForegroundColor Green
Write-Host "Run 'springcli' to get started." -ForegroundColor Cyan
