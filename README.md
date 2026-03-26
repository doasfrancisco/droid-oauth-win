<p align="center"><img src="droid_proxy_windows.png" width="120"></p>

<h1 align="center">Droid OAuth Windows</h1>

<p align="center">Use your ChatGPT Plus/Pro subscription with Factory Droid on Windows via OAuth proxy.</p>

## Install

```powershell
irm https://raw.githubusercontent.com/doasfrancisco/droid-oauth-win/main/install.ps1 | iex
```

## How it works

```
Factory Droid → http://127.0.0.1:8317/v1 → [OAuth tokens] → OpenAI API
```

One exe. Double-click it. First launch opens your browser for OpenAI sign-in. After that, it sits in the system tray and proxies Droid requests through your subscription.

Droid's `settings.json` is auto-configured with GPT-5.4 and GPT-5.3-Codex models pointing at the local proxy.

## Install (pre-built)

1. Download `DroidOAuthWindows.exe` from [Releases](../../releases)
2. Double-click
3. Sign in with your OpenAI account (first time only)
4. Open Droid, select **GPT-5.4 [Plus]** or **GPT-5.3-Codex [Plus]**

## Build from source

### Prerequisites

- [Go 1.26+](https://golang.org/dl/) (auto-downloads via toolchain if you have Go 1.24+)
- [Git](https://git-scm.com/)

### Build

```powershell
git clone https://github.com/doasfrancisco/droid-oauth-win.git
cd droid-oauth-win
go mod tidy
go build -ldflags "-H windowsgui" -o DroidOAuthWindows.exe .
```

The `-H windowsgui` flag makes it a GUI app — no console window on launch.

### Modify the exe

The entire app is `main.go`. Here's what each part does:

| Section | What it does |
|---|---|
| `buildConfig()` | Proxy server config (port, auth dir, retries). Change `proxyPort` to use a different port. |
| `ensureDroidConfig()` | Writes custom models to `~/.factory/settings.json`. Add/remove models here. |
| `doLogin()` | OpenAI OAuth flow. Uses CLIProxyAPI's SDK — opens browser, catches callback, saves token. |
| `startProxy()` | Starts the proxy in-process via CLIProxyAPI SDK. Checks for tokens first, triggers login if missing. |
| `onReady()` | System tray setup — icon, menu items, event loop. |
| `icon.ico` | Tray icon — embedded at compile time via `//go:embed`. Replace the file and rebuild. |

After changes:

```powershell
go build -ldflags "-H windowsgui" -o DroidOAuthWindows.exe .
```

### Add more models

Edit the `proxyModels` slice in `ensureDroidConfig()`:

```go
{
    "model":       "o3",
    "id":          "custom:o3-[Plus]-2",
    "index":       2,
    "baseUrl":     fmt.Sprintf("http://127.0.0.1:%d/v1", proxyPort),
    "apiKey":      "dummy-not-used",
    "displayName": "o3 [Plus]",
    "provider":    "openai",
},
```

Rebuild and re-run. The new model appears in Droid's model selector.

## Tray menu

Right-click the tray icon:

- **Running on 127.0.0.1:8317** — status (disabled, info only)
- **Re-login OpenAI** — re-authenticate if tokens expire
- **Restart** — restart the proxy
- **Stop & Exit** — kill proxy and quit

## Token storage

OAuth tokens are saved to `%USERPROFILE%\.cli-proxy-api\`. Tokens auto-refresh. If you get auth errors, use **Re-login OpenAI** from the tray menu.

## Known issues

- **`/compress` breaks** through the proxy — switch to a native Factory model, run `/compact`, switch back
- **Tool calling** can be flaky with some models through the proxy
- Uses `127.0.0.1` instead of `localhost` — Factory filters out `localhost` URLs

## Credits

Built on [CLIProxyAPI](https://github.com/luispater/CLIProxyAPI) by luispater.
