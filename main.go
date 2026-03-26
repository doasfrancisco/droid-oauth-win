package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getlantern/systray"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/config"
)

//go:embed icon.ico
var appIcon []byte

var (
	cancel  context.CancelFunc
	mu      sync.Mutex
	stopped bool
)

const proxyPort = 8317

func main() {
	ensureDroidConfig()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(appIcon)
	systray.SetTooltip("Droid OAuth Win - Starting...")

	mStatus := systray.AddMenuItem("Starting...", "Proxy status")
	mStatus.Disable()
	systray.AddSeparator()
	mLogin := systray.AddMenuItem("Re-login OpenAI", "Re-authenticate with OpenAI")
	mRestart := systray.AddMenuItem("Restart", "Restart the proxy")
	mQuit := systray.AddMenuItem("Stop && Exit", "Stop the proxy and exit")

	go startProxy(mStatus)

	go func() {
		for {
			select {
			case <-mLogin.ClickedCh:
				doLogin()
			case <-mRestart.ClickedCh:
				stopProxy()
				mStatus.SetTitle("Restarting...")
				systray.SetTooltip("Droid OAuth Win - Restarting...")
				go startProxy(mStatus)
			case <-mQuit.ClickedCh:
				stopProxy()
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	stopProxy()
}

func buildConfig() (*config.Config, string) {
	home, _ := os.UserHomeDir()
	cfgDir := filepath.Join(home, ".cli-proxy-api")
	os.MkdirAll(cfgDir, 0700)
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// Write config file if it doesn't exist
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgData := fmt.Sprintf("host: \"127.0.0.1\"\nport: %d\nauth-dir: \"%s\"\nrequest-retry: 3\nusage-statistics-enabled: true\n", proxyPort, cfgDir)
		os.WriteFile(cfgPath, []byte(cfgData), 0644)
	}

	cfg, _ := config.LoadConfig(cfgPath)
	if cfg == nil {
		cfg = &config.Config{
			Host:                   "127.0.0.1",
			Port:                   proxyPort,
			AuthDir:                cfgDir,
			UsageStatisticsEnabled: true,
			RequestRetry:           3,
		}
	}
	return cfg, cfgPath
}

func hasTokens() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	dir := filepath.Join(home, ".cli-proxy-api")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".json" {
			return true
		}
	}
	return false
}

func doLogin() {
	cfg, _ := buildConfig()
	store := auth.NewFileTokenStore()

	home, _ := os.UserHomeDir()
	authDir := filepath.Join(home, ".cli-proxy-api")
	os.MkdirAll(authDir, 0700)
	store.SetBaseDir(authDir)

	mgr := auth.NewManager(store, auth.NewCodexAuthenticator())
	_, _, err := mgr.Login(context.Background(), "codex", cfg, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
	}
}

func startProxy(status *systray.MenuItem) {
	if !hasTokens() {
		status.SetTitle("No tokens - logging in...")
		systray.SetTooltip("Droid OAuth Win - Logging in...")
		doLogin()
		if !hasTokens() {
			status.SetTitle("Login failed - right-click to retry")
			systray.SetTooltip("Droid OAuth Win - No tokens")
			return
		}
	}

	cfg, cfgPath := buildConfig()

	svc, err := cliproxy.NewBuilder().
		WithConfig(cfg).
		WithConfigPath(cfgPath).
		Build()
	if err != nil {
		logErr(fmt.Sprintf("Build error: %v", err))
		status.SetTitle(fmt.Sprintf("Error: %v", err))
		systray.SetTooltip("Droid OAuth Win - Error")
		return
	}

	mu.Lock()
	ctx, c := context.WithCancel(context.Background())
	cancel = c
	stopped = false
	mu.Unlock()

	status.SetTitle(fmt.Sprintf("Running on 127.0.0.1:%d", proxyPort))
	systray.SetTooltip(fmt.Sprintf("Droid OAuth Win - Running on 127.0.0.1:%d", proxyPort))

	err = svc.Run(ctx)

	mu.Lock()
	intentional := stopped
	mu.Unlock()

	if !intentional && err != nil {
		logErr(fmt.Sprintf("Run error: %v", err))
		status.SetTitle("Proxy stopped unexpectedly")
		systray.SetTooltip("Droid OAuth Win - Stopped")
	}
}

func stopProxy() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		stopped = true
		cancel()
		cancel = nil
	}
}

func logErr(msg string) {
	home, _ := os.UserHomeDir()
	f, err := os.OpenFile(filepath.Join(home, ".cli-proxy-api", "droid-oauth-win.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, msg)
}

// ensureDroidConfig writes custom models to Factory Droid's settings.json if not already present.
func ensureDroidConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	settingsPath := filepath.Join(home, ".factory", "settings.json")

	var settings map[string]interface{}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// No settings file — create .factory dir and start fresh
		os.MkdirAll(filepath.Join(home, ".factory"), 0755)
		settings = map[string]interface{}{}
	} else {
		json.Unmarshal(data, &settings)
	}

	// Check if customModels already has our proxy entries
	if models, ok := settings["customModels"]; ok {
		if arr, ok := models.([]interface{}); ok {
			for _, m := range arr {
				if obj, ok := m.(map[string]interface{}); ok {
					if url, _ := obj["baseUrl"].(string); url == fmt.Sprintf("http://127.0.0.1:%d/v1", proxyPort) {
						return // already configured
					}
				}
			}
		}
	}

	// Add our models
	proxyModels := []map[string]interface{}{
		{
			"model":          "gpt-5.4",
			"id":             "custom:GPT-5.4-[Plus]-0",
			"index":          0,
			"baseUrl":        fmt.Sprintf("http://127.0.0.1:%d/v1", proxyPort),
			"apiKey":         "dummy-not-used",
			"displayName":    "GPT-5.4 [Plus]",
			"noImageSupport": false,
			"provider":       "openai",
		},
		{
			"model":          "gpt-5.3-codex",
			"id":             "custom:GPT-5.3-Codex-[Plus]-1",
			"index":          1,
			"baseUrl":        fmt.Sprintf("http://127.0.0.1:%d/v1", proxyPort),
			"apiKey":         "dummy-not-used",
			"displayName":    "GPT-5.3-Codex [Plus]",
			"noImageSupport": false,
			"provider":       "openai",
		},
	}

	existing, _ := settings["customModels"].([]interface{})
	for _, pm := range proxyModels {
		existing = append(existing, pm)
	}
	settings["customModels"] = existing

	out, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(settingsPath, out, 0644)
}
