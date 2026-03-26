$ErrorActionPreference = "Stop"

$repo = "doasfrancisco/droid-oauth-win"
$asset = "DroidOAuthWindows.exe"
$installDir = "$env:LOCALAPPDATA\DroidOAuth"
$exePath = "$installDir\$asset"
$shortcutName = "Droid OAuth Windows"

# Uninstall mode
if ($args -contains "--uninstall" -or $args -contains "-u") {
    Write-Host "Uninstalling Droid OAuth Windows..." -ForegroundColor Cyan
    # Kill if running
    Get-Process -Name "DroidOAuthWindows" -ErrorAction SilentlyContinue | Stop-Process -Force
    # Remove shortcuts
    Remove-Item "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\$shortcutName.lnk" -ErrorAction SilentlyContinue
    Remove-Item "$env:USERPROFILE\Desktop\$shortcutName.lnk" -ErrorAction SilentlyContinue
    # Remove install dir
    Remove-Item $installDir -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "Uninstalled." -ForegroundColor Green
    exit 0
}

Write-Host "Installing Droid OAuth Windows..." -ForegroundColor Cyan

# Get latest release download URL
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$url = ($release.assets | Where-Object { $_.name -eq $asset }).browser_download_url

if (-not $url) {
    Write-Host "ERROR: Could not find $asset in latest release." -ForegroundColor Red
    exit 1
}

# Create install directory
New-Item -ItemType Directory -Path $installDir -Force | Out-Null

# Download
Write-Host "Downloading from $url..."
Invoke-WebRequest -Uri $url -OutFile $exePath -UseBasicParsing

# Create shortcuts
$shell = New-Object -ComObject WScript.Shell

# Start Menu
$startMenu = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs"
$shortcut = $shell.CreateShortcut("$startMenu\$shortcutName.lnk")
$shortcut.TargetPath = $exePath
$shortcut.WorkingDirectory = $installDir
$shortcut.Description = "Use ChatGPT Plus/Pro with Factory Droid via OAuth proxy"
$shortcut.Save()

# Desktop
$desktop = "$env:USERPROFILE\Desktop"
$shortcut = $shell.CreateShortcut("$desktop\$shortcutName.lnk")
$shortcut.TargetPath = $exePath
$shortcut.WorkingDirectory = $installDir
$shortcut.Description = "Use ChatGPT Plus/Pro with Factory Droid via OAuth proxy"
$shortcut.Save()

Write-Host ""
Write-Host "Installed to: $exePath" -ForegroundColor Green
Write-Host "Shortcuts: Start Menu + Desktop" -ForegroundColor Green
Write-Host ""
Write-Host "To uninstall:" -ForegroundColor Yellow
Write-Host "  irm https://raw.githubusercontent.com/$repo/main/install.ps1 | iex -u"
