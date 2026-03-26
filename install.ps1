$ErrorActionPreference = "Stop"

$repo = "doasfrancisco/droid-oauth-win"
$asset = "DroidOAuthWindows.exe"
$installDir = "$env:LOCALAPPDATA\DroidOAuth"
$exePath = "$installDir\$asset"
$shortcutName = "Droid OAuth Windows"

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

# Create Start Menu shortcut
$startMenu = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs"
$shell = New-Object -ComObject WScript.Shell
$shortcut = $shell.CreateShortcut("$startMenu\$shortcutName.lnk")
$shortcut.TargetPath = $exePath
$shortcut.WorkingDirectory = $installDir
$shortcut.Description = "Use ChatGPT Plus/Pro with Factory Droid via OAuth proxy"
$shortcut.Save()

Write-Host ""
Write-Host "Installed to: $exePath" -ForegroundColor Green
Write-Host "Start Menu shortcut: $shortcutName" -ForegroundColor Green
Write-Host ""
Write-Host "Launch it from Start Menu or run:" -ForegroundColor Cyan
Write-Host "  & '$exePath'"
